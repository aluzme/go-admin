package controller

import (
	"fmt"
	template2 "html/template"
	"net/http"
	"net/url"

	"github.com/aluzme/go-admin/modules/logger"

	"github.com/aluzme/go-admin/template"

	"github.com/aluzme/go-admin/plugins/admin/modules/response"

	"github.com/aluzme/go-admin/context"
	"github.com/aluzme/go-admin/modules/auth"
	"github.com/aluzme/go-admin/modules/file"
	"github.com/aluzme/go-admin/modules/language"
	"github.com/aluzme/go-admin/plugins/admin/modules"
	"github.com/aluzme/go-admin/plugins/admin/modules/constant"
	form2 "github.com/aluzme/go-admin/plugins/admin/modules/form"
	"github.com/aluzme/go-admin/plugins/admin/modules/guard"
	"github.com/aluzme/go-admin/plugins/admin/modules/parameter"
	"github.com/aluzme/go-admin/template/types"
	"github.com/aluzme/go-admin/template/types/form"
)

// ShowForm show form page.
func (h *Handler) ShowForm(ctx *context.Context) {
	param := guard.GetShowFormParam(ctx)
	h.showForm(ctx, "", param.Prefix, param.Param, false)
}

func (h *Handler) showForm(ctx *context.Context, alert template2.HTML, prefix string, param parameter.Parameters, isEdit bool, animation ...bool) {

	panel := h.table(prefix, ctx)

	if panel.GetForm().HasError() {
		if panel.GetForm().PageErrorHTML != template2.HTML("") {
			h.HTML(ctx, auth.Auth(ctx),
				types.Panel{Content: panel.GetForm().PageErrorHTML}, param.Animation)
			return
		}
		h.HTML(ctx, auth.Auth(ctx),
			template.WarningPanel(panel.GetForm().PageError.Error(),
				template.GetPageTypeFromPageError(panel.GetForm().PageError)), param.Animation)
		return
	}

	user := auth.Auth(ctx)

	paramStr := param.GetRouteParamStr()

	newUrl := modules.AorEmpty(panel.GetCanAdd(), h.routePathWithPrefix("show_new", prefix)+paramStr)
	footerKind := "edit"
	if newUrl == "" || !user.CheckPermissionByUrlMethod(newUrl, h.route("show_new").Method(), url.Values{}) {
		footerKind = "edit_only"
	}

	formInfo, err := panel.GetDataWithId(param)

	if err != nil {
		logger.Error("receive data error: ", err)
		h.HTML(ctx, user, template.
			WarningPanelWithDescAndTitle(err.Error(), panel.GetForm().Description, panel.GetForm().Title),
			alert == "" || ((len(animation) > 0) && animation[0]))

		if isEdit {
			showEditUrl := h.routePathWithPrefix("show_edit", prefix) + param.DeletePK().GetRouteParamStr()
			ctx.AddHeader(constant.PjaxUrlHeader, showEditUrl)
		}
		return
	}

	infoUrl := h.routePathWithPrefix("info", prefix) + param.DeleteField(constant.EditPKKey).GetRouteParamStr()
	editUrl := h.routePathWithPrefix("edit", prefix)

	referer := ctx.Headers("Referer")

	if referer != "" && !isInfoUrl(referer) && !isEditUrl(referer, ctx.Query(constant.PrefixKey)) {
		infoUrl = referer
	}

	f := panel.GetForm()

	isNotIframe := ctx.Query(constant.IframeKey) != "true"

	hiddenFields := map[string]string{
		form2.TokenKey:    h.authSrv().AddToken(),
		form2.PreviousKey: infoUrl,
	}

	if ctx.Query(constant.IframeKey) != "" {
		hiddenFields[constant.IframeKey] = ctx.Query(constant.IframeKey)
	}

	if ctx.Query(constant.IframeIDKey) != "" {
		hiddenFields[constant.IframeIDKey] = ctx.Query(constant.IframeIDKey)
	}

	content := formContent(aForm().
		SetContent(formInfo.FieldList).
		SetFieldsHTML(f.HTMLContent).
		SetTabContents(formInfo.GroupFieldList).
		SetTabHeaders(formInfo.GroupFieldHeaders).
		SetPrefix(h.config.PrefixFixSlash()).
		SetInputWidth(f.InputWidth).
		SetHeadWidth(f.HeadWidth).
		SetPrimaryKey(panel.GetPrimaryKey().Name).
		SetUrl(editUrl).
		SetTitle(f.FormEditTitle).
		SetAjax(f.AjaxSuccessJS, f.AjaxErrorJS).
		SetLayout(f.Layout).
		SetHiddenFields(hiddenFields).
		SetOperationFooter(formFooter(footerKind,
			f.IsHideContinueEditCheckBox,
			f.IsHideContinueNewCheckBox,
			f.IsHideResetButton, f.FormEditBtnWord)).
		SetHeader(f.HeaderHtml).
		SetFooter(f.FooterHtml), len(formInfo.GroupFieldHeaders) > 0, !isNotIframe, f.IsHideBackButton, f.Header)

	if f.Wrapper != nil {
		content = f.Wrapper(content)
	}

	h.HTML(ctx, user, types.Panel{
		Content:     alert + content,
		Description: template2.HTML(formInfo.Description),
		Title:       modules.AorBHTML(isNotIframe, template2.HTML(formInfo.Title), ""),
	}, alert == "" || ((len(animation) > 0) && animation[0]))

	if isEdit {
		showEditUrl := h.routePathWithPrefix("show_edit", prefix) + param.DeletePK().GetRouteParamStr()
		ctx.AddHeader(constant.PjaxUrlHeader, showEditUrl)
	}
}

func (h *Handler) EditForm(ctx *context.Context) {

	param := guard.GetEditFormParam(ctx)

	if len(param.MultiForm.File) > 0 {
		err := file.GetFileEngine(h.config.FileUploadEngine.Name).Upload(param.MultiForm)
		if err != nil {
			logger.Error("get file engine error: ", err)
			if ctx.WantJSON() {
				response.Error(ctx, err.Error())
			} else {
				h.showForm(ctx, aAlert().Warning(err.Error()), param.Prefix, param.Param, true)
			}
			return
		}
	}

	for _, field := range param.Panel.GetForm().FieldList {
		if field.FormType == form.File &&
			len(param.MultiForm.File[field.Field]) == 0 &&
			param.MultiForm.Value[field.Field+"__delete_flag"][0] != "1" {
			delete(param.MultiForm.Value, field.Field)
		}
	}

	err := param.Panel.UpdateData(param.Value())
	if err != nil {
		logger.Error("update data error: ", err)
		if ctx.WantJSON() {
			response.Error(ctx, err.Error())
		} else {
			h.showForm(ctx, aAlert().Warning(err.Error()), param.Prefix, param.Param, true)
		}
		return
	}

	if param.Panel.GetForm().Responder != nil {
		param.Panel.GetForm().Responder(ctx)
		return
	}

	if ctx.WantJSON() && !param.IsIframe {
		response.OkWithData(ctx, map[string]interface{}{
			"url": param.PreviousPath,
		})
		return
	}

	if !param.FromList {

		if isNewUrl(param.PreviousPath, param.Prefix) {
			h.showNewForm(ctx, param.Alert, param.Prefix, param.Param.DeleteEditPk().GetRouteParamStr(), true)
			return
		}

		if isEditUrl(param.PreviousPath, param.Prefix) {
			h.showForm(ctx, param.Alert, param.Prefix, param.Param, true, false)
			return
		}

		ctx.HTML(http.StatusOK, fmt.Sprintf(`<script>location.href="%s"</script>`, param.PreviousPath))
		ctx.AddHeader(constant.PjaxUrlHeader, param.PreviousPath)
		return
	}

	if param.IsIframe {
		ctx.HTML(http.StatusOK, fmt.Sprintf(`<script>
		swal('%s', '', 'success');
		setTimeout(function(){
			$("#%s", window.parent.document).hide();
			$('.modal-backdrop.fade.in', window.parent.document).hide();
		}, 1000)
</script>`, language.Get("success"), param.IframeID))
		return
	}

	buf := h.showTable(ctx, param.Prefix, param.Param.DeletePK().DeleteEditPk(), nil)

	ctx.HTML(http.StatusOK, buf.String())
	ctx.AddHeader(constant.PjaxUrlHeader, param.PreviousPath)
}

package main

import (
	"fmt"
	"net/http"
)

// TODO 通用错误error处理包
// logError 方法是用于在应用程序中记录错误信息的通用帮助函数，同时
// 记录请求的方法和请求的URL。
func (app *application) logError(r *http.Request, err error) {
	app.logger.PrintError(err, map[string]string{
		"request_method": r.Method,
		"request_url":    r.URL.String(),
	})
	//app.logger.Println(err)
}

// errorResponse 方法是用于向客户端发送JSON格式错误消息的通用帮助函数。
// 注意，我们使用 interface{} 类型而不是字符串类型，这使得我们可以更灵活地处理响应中的值。
func (app *application) errorResponse(w http.ResponseWriter, r *http.Request, status int, message interface{}) {
	env := envelope{"error": message}

	// 使用 writeJSON() 帮助函数写入响应。如果返回错误，则记录错误，并向客户端发送空的500内部服务器错误状态码。
	err := app.writeJSON(w, status, env, nil)
	if err != nil {
		app.logError(r, err)
		w.WriteHeader(500)
	}
}

// serverErrorResponse 方法用于处理应用程序在运行时遇到的意外问题。
// 它记录详细的错误信息，并使用 errorResponse() 帮助函数向客户端发送500内部服务器错误状态码和JSON响应。
func (app *application) serverErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logError(r, err)

	message := "the server encountered a problem and could not process your request"
	app.errorResponse(w, r, 500, message)
}

// notFoundResponse 方法用于向客户端发送404未找到状态码和JSON响应。
func (app *application) notFoundResponse(w http.ResponseWriter, r *http.Request) {
	message := "the requested resource could not be found"
	app.errorResponse(w, r, http.StatusNotFound, message)
}

// methodNotAllowedResponse 方法用于向客户端发送405方法不允许状态码和JSON响应。
func (app *application) methodNotAllowedResponse(w http.ResponseWriter, r *http.Request) {
	message := fmt.Sprintf("the %s method is not supported this resource", r.Method)
	app.errorResponse(w, r, http.StatusMethodNotAllowed, message)
}

// badRequestResponse 向客户端发送400错误请求状态码和JSON格式的错误消息。
func (app *application) badRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.errorResponse(w, r, http.StatusBadRequest, err.Error())
}

// failedValidationResponse 在验证失败时向客户端发送422不可处理实体状态码和JSON格式的错误消息。
// 注意，errors 参数的类型是 map[string]string，这与我们的 Validator 类型中的 errors 映射完全相同。
func (app *application) failedValidationResponse(w http.ResponseWriter, r *http.Request, errors map[string]string) {
	app.errorResponse(w, r, http.StatusUnprocessableEntity, errors)
}

// editConflictResponse 向客户端发送409冲突状态码和JSON格式的错误消息。
func (app *application) editConflictResponse(w http.ResponseWriter, r *http.Request) {
	message := "unable to update the record due to an edit conflict, please try again"
	app.errorResponse(w, r, http.StatusConflict, message)
}

// rateLimitExceedResponse 向客户端发送429太多请求状态码和JSON格式的错误消息。
func (app *application) rateLimitExceededResponse(w http.ResponseWriter, r *http.Request) {
	message := "rate limited exceeded"
	app.errorResponse(w, r, http.StatusTooManyRequests, message)
}

// invalidCredentialsResponse 向客户端发送401未授权状态码和JSON格式的错误消息。
func (app *application) invalidCredentialsResponse(w http.ResponseWriter, r *http.Request) {
	message := "invalid authentication credentials"
	app.errorResponse(w, r, http.StatusUnauthorized, message)
}

// invalidAuthenticationTokenResponse 向客户端发送401未授权状态码和"WWW-Authenticate: Bearer"头以及JSON格式的错误消息。
func (app *application) invalidAuthenticationTokenResponse(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("WWWW-Authenticate", "Bearer")

	message := "invalid or missing authentication token"
	app.errorResponse(w, r, http.StatusUnauthorized, message)
}

// authenticationRequiredResponse 向客户端发送401未授权状态码和JSON格式的错误消息。
func (app *application) authenticationRequiredResponse(w http.ResponseWriter, r *http.Request) {
	message := "you must be authenticated to access this resource"
	app.errorResponse(w, r, http.StatusUnauthorized, message)
}

// inactiveAccountResponse 向客户端发送403禁止状态码和JSON格式的错误消息。
func (app *application) inactiveAccountResponse(w http.ResponseWriter, r *http.Request) {
	message := "your user account must be activated to access this resource"
	app.errorResponse(w, r, http.StatusForbidden, message)
}

// notPermittedResponse 向客户端发送403禁止状态码和JSON格式的错误消息。
// 注意，此帮助函数用于处理没有权限访问资源的用户。
func (app *application) notPermittedResponse(w http.ResponseWriter, r *http.Request) {
	message := "your user account doesn't have the necessary permissions to access this resource"
	app.errorResponse(w, r, http.StatusForbidden, message)
}

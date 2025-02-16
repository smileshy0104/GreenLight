package main

import (
	"DesignMode/GreenLight/internal/data"
	"DesignMode/GreenLight/internal/validator"
	"errors"
	"net/http"
)

// 注册用户 registerUserHandler
func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	// 声明结构体 input
	var input struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	// 读取JSON请求体数据到input结构体中
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user := &data.User{
		Name:      input.Name,
		Email:     input.Email,
		Activated: false,
	}

	// 使用 Password.Set() 方法设置密码（加密）
	err = user.Password.Set(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	v := validator.New()

	// 使用 ValidateUser() 方法对用户数据进行验证
	if data.ValidateUser(v, user); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// 调用 Users.Insert() 方法将用户插入数据库
	err = app.models.Users.Insert(user)
	if err != nil {
		switch {
		// 如果是 ErrDuplicateEmail，则说明该邮箱已经被注册，因此返回一个错误响应
		case errors.Is(err, data.ErrDuplicateEmail):
			v.AddError("email", "a user with this email address already exists")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// TODO 发送激活邮件相关代码
	// ‘background’ goroutine将发送欢迎邮件，这个代码将被并发执行，我们不需在等待邮件发送！
	//go func() {
	//	// 添加defer语句，防止程序崩溃时忘记关闭资源
	//	defer func() {
	//		if err := recover(); err != nil {
	//			app.logger.PrintError(fmt.Errorf("%s", err), nil)
	//		}
	//	}()
	//	err = app.mailer.Send(user.Email, "user_welcome.tmpl", user)
	//	if err != nil {
	//		app.logger.PrintError(err, nil)
	//	}
	//}()

	// TODO 直接调用封装的background()函数，来捕获程序崩溃
	app.background(func() {
		err = app.mailer.Send(user.Email, "user_welcome.tmpl", user)
		if err != nil {
			app.logger.PrintError(err, nil)
		}
	})

	err = app.writeJSON(w, http.StatusAccepted, envelope{"user": user}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// activateUserHandler activates a user by setting 'activation = true' using the provided
// activation token in the request body.
func (app *application) activateUserHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the plaintext activation token from the request body
	var input struct {
		TokenPlaintext string `json:"token"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// Validate the plaintext token provided by the client.
	v := validator.New()

	//if data.ValidateTokenPlaintext(v, input.TokenPlaintext); !v.Valid() {
	//	app.failedValidationResponse(w, r, v.Errors)
	//	return
	//}

	// Retrieve the details of the user associated with the token using the GetForToken() method.
	// If no matching record is found, then we let the client know that the token they provided
	// is not valid.
	user, err := app.models.Users.GetForToken(data.ScopeActivation, input.TokenPlaintext)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			v.AddError("token", "invalid or expired activation token")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// Update the user's activation status.
	user.Activated = true

	// Save the updated user record in our database, checking for any edit conflicts in the same
	// way that we did for our move records.
	err = app.models.Users.Update(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// If everything went successfully above, then delete all activation tokens for the user.
	err = app.models.Tokens.DeleteAllForUser(data.ScopeActivation, user.ID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"user": user}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

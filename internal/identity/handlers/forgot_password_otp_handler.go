package handler

import (
	"errors"
	"identityservice/internal/httpresponse"
	"identityservice/internal/identity/dto/request"
	"identityservice/internal/identity/dto/response"
	"identityservice/internal/identity/usecases"
	"log"
	"net/http"

	"github.com/labstack/echo/v5"
)

type ForgotPasswordOTPHandler struct {
	requestOTPUseCase usecases.RequestForgotPasswordOTPUseCase
	verifyOTPUseCase  usecases.VerifyForgotPasswordOTPUseCase
}

func NewForgotPasswordOTPHandler(requestOTPUseCase usecases.RequestForgotPasswordOTPUseCase, verifyOTPUseCase usecases.VerifyForgotPasswordOTPUseCase) *ForgotPasswordOTPHandler {
	return &ForgotPasswordOTPHandler{requestOTPUseCase: requestOTPUseCase, verifyOTPUseCase: verifyOTPUseCase}
}

func (h ForgotPasswordOTPHandler) Request(c *echo.Context) error {
	payload := new(request.RequestForgotPasswordOTPRequest)
	if err := c.Bind(payload); err != nil {
		return c.JSON(http.StatusBadRequest, httpresponse.Error("invalid payload"))
	}
	if err := c.Validate(payload); err != nil {
		return c.JSON(http.StatusBadRequest, httpresponse.Error("payload validation failed"))
	}
	if err := h.requestOTPUseCase.Request(c.Request().Context(), *payload); err != nil {
		return writeForgotPasswordOTPError(c, err)
	}
	return c.JSON(http.StatusOK, httpresponse.Success(nil, "forgot password OTP sent"))
}

func (h ForgotPasswordOTPHandler) Verify(c *echo.Context) error {
	payload := new(request.VerifyForgotPasswordOTPRequest)
	if err := c.Bind(payload); err != nil {
		return c.JSON(http.StatusBadRequest, httpresponse.Error("invalid payload"))
	}
	if err := c.Validate(payload); err != nil {
		return c.JSON(http.StatusBadRequest, httpresponse.Error("payload validation failed"))
	}
	token, expiresIn, err := h.verifyOTPUseCase.Verify(c.Request().Context(), *payload)
	if err != nil {
		return writeForgotPasswordOTPError(c, err)
	}
	return c.JSON(http.StatusOK, httpresponse.Success(response.VerifyRegisterOTPResponse{RegistrationToken: token, ExpiresIn: expiresIn}, "forgot password OTP verified"))
}

func writeForgotPasswordOTPError(c *echo.Context, err error) error {
	switch {
	case errors.Is(err, usecases.ErrUserNotFound):
		return c.JSON(http.StatusBadRequest, httpresponse.Error("no account found for this email"))
	case errors.Is(err, usecases.ErrDifferentAuthProvider):
		return c.JSON(http.StatusBadRequest, httpresponse.Error("password reset is only available for local accounts"))
	case errors.Is(err, usecases.ErrSmtpError):
		log.Printf("[IDENTITY] ForgotPasswordOTP email error: %v", err)
		return c.JSON(http.StatusInternalServerError, httpresponse.Error("failed to send password reset OTP email"))
	default:
		if errors.Is(err, usecases.ErrInternalError) {
			log.Printf("[IDENTITY] ForgotPasswordOTP error: %v", err)
		}
		return writeRegisterOTPError(c, err)
	}
}

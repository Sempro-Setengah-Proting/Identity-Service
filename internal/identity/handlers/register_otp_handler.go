package handler

import (
	"errors"
	"identityservice/internal/httpresponse"
	"identityservice/internal/identity/dto/request"
	"identityservice/internal/identity/dto/response"
	"identityservice/internal/identity/usecases"
	"net/http"

	"github.com/labstack/echo/v5"
)

type RegisterOTPHandler struct {
	requestOTPUseCase usecases.RequestRegisterOTPUseCase
	verifyOTPUseCase  usecases.VerifyRegisterOTPUseCase
}

func NewRegisterOTPHandler(
	requestOTPUseCase usecases.RequestRegisterOTPUseCase,
	verifyOTPUseCase usecases.VerifyRegisterOTPUseCase,
) *RegisterOTPHandler {
	return &RegisterOTPHandler{
		requestOTPUseCase: requestOTPUseCase,
		verifyOTPUseCase:  verifyOTPUseCase,
	}
}

func (h RegisterOTPHandler) Request(c *echo.Context) error {
	ctx := c.Request().Context()
	payload := new(request.RequestRegisterOTPRequest)

	if err := c.Bind(payload); err != nil {
		return c.JSON(http.StatusBadRequest, httpresponse.Error("invalid payload"))
	}
	if err := c.Validate(payload); err != nil {
		return c.JSON(http.StatusBadRequest, httpresponse.Error("payload validation failed"))
	}

	if err := h.requestOTPUseCase.Request(ctx, *payload); err != nil {
		return writeRegisterOTPError(c, err)
	}

	return c.JSON(http.StatusOK, httpresponse.Success(nil, "register OTP sent"))
}

func (h RegisterOTPHandler) Verify(c *echo.Context) error {
	ctx := c.Request().Context()
	payload := new(request.VerifyRegisterOTPRequest)

	if err := c.Bind(payload); err != nil {
		return c.JSON(http.StatusBadRequest, httpresponse.Error("invalid payload"))
	}
	if err := c.Validate(payload); err != nil {
		return c.JSON(http.StatusBadRequest, httpresponse.Error("payload validation failed"))
	}

	registrationToken, expiresIn, err := h.verifyOTPUseCase.Verify(ctx, *payload)
	if err != nil {
		return writeRegisterOTPError(c, err)
	}

	return c.JSON(http.StatusOK, httpresponse.Success(response.VerifyRegisterOTPResponse{
		RegistrationToken: registrationToken,
		ExpiresIn:         expiresIn,
	}, "register OTP verified"))
}

func writeRegisterOTPError(c *echo.Context, err error) error {
	switch {
	case errors.Is(err, usecases.ErrEmailAlreadyRegistered):
		return c.JSON(http.StatusConflict, httpresponse.Error("email already registered"))
	case errors.Is(err, usecases.ErrOTPResendCooldown):
		return c.JSON(http.StatusTooManyRequests, httpresponse.Error("please wait before requesting another OTP"))
	case errors.Is(err, usecases.ErrOTPMaxAttempts):
		return c.JSON(http.StatusTooManyRequests, httpresponse.Error("maximum OTP attempts reached"))
	case errors.Is(err, usecases.ErrOTPNotFound):
		return c.JSON(http.StatusBadRequest, httpresponse.Error("OTP not found or expired"))
	case errors.Is(err, usecases.ErrOTPInvalid):
		return c.JSON(http.StatusBadRequest, httpresponse.Error("invalid OTP"))
	default:
		return c.JSON(http.StatusInternalServerError, httpresponse.Error("internal error"))
	}
}

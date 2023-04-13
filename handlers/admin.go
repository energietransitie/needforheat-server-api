package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/energietransitie/twomes-backoffice-api/pkg/ports"
	"github.com/energietransitie/twomes-backoffice-api/pkg/twomes"
	"github.com/sirupsen/logrus"
)

// AdminHandler can be used in an RPC server.
// It also has an HTTP middleware to verify admin tokens with admin accounts.
type AdminHandler struct {
	service ports.AdminService
}

func NewAdminHandler(service ports.AdminService) *AdminHandler {
	return &AdminHandler{
		service: service,
	}
}

func (h *AdminHandler) List(input int, reply *[]twomes.Admin) error {
	admins, err := h.service.GetAll()
	if err != nil {
		return err
	}

	*reply = admins
	return nil
}

func (h *AdminHandler) Create(admin twomes.Admin, token *string) error {
	admin, err := h.service.Create(admin.Name, admin.Expiry)
	if err != nil {
		return err
	}

	*token = admin.AuthorizationToken
	return nil
}

func (h *AdminHandler) Delete(admin twomes.Admin, reply *twomes.Admin) error {
	return h.service.Delete(admin)
}

func (h *AdminHandler) Reactivate(admin twomes.Admin, reply *twomes.Admin) error {
	admin, err := h.service.Reactivate(admin)
	if err != nil {
		return err
	}

	*reply = admin
	return nil
}

func (h *AdminHandler) SetExpiry(admin twomes.Admin, reply *twomes.Admin) error {
	expiry := admin.Expiry

	admin.Expiry = time.Time{}

	admin, err := h.service.SetExpiry(admin, expiry)
	if err != nil {
		return err
	}

	*reply = admin
	return nil
}

// HTTP middleware to check if admin in admin auth token is valid.
func (h *AdminHandler) Middleware(next Handler) Handler {
	return func(w http.ResponseWriter, r *http.Request) error {
		auth, ok := r.Context().Value(AuthorizationCtxKey).(*twomes.Authorization)
		if !ok {
			return NewHandlerError(nil, "unauthorized", http.StatusUnauthorized).WithMessage("failed when getting authentication context value")
		}

		admin, err := h.service.Find(twomes.Admin{ID: auth.ID})
		if err != nil {
			return NewHandlerError(err, "forbidden", http.StatusForbidden).WithMessage("failed matching admin to auth details")
		}

		if auth.Claims.IssuedAt.Before(admin.ActivatedAt) {
			return NewHandlerError(err, "forbidden", http.StatusForbidden).WithMessage(fmt.Sprintf("admin \"%s\" tried to use an invalidated token", admin.Name)).WithLevel(logrus.WarnLevel)
		}

		if admin.Expiry.Before(time.Now()) {
			return NewHandlerError(err, "forbidden", http.StatusForbidden).WithMessage(fmt.Sprintf("token for expired admin \"%s\" was used", admin.Name)).WithLevel(logrus.WarnLevel)
		}

		return next(w, r)
	}
}

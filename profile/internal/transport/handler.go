package transport

import (
	"errors"
	"net/http"

	"github.com/TATAROmangol/mess/profile/internal/domain"
	"github.com/TATAROmangol/mess/profile/internal/model"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	domain domain.Service
}

func NewHandler(domain domain.Service) *Handler {
	return &Handler{
		domain: domain,
	}
}

func (h *Handler) GetProfile(c *gin.Context) {
	id := c.Param("id")

	var profile *model.Profile
	var url string
	var err error

	if id == "" {
		profile, url, err = h.domain.GetCurrentProfile(c.Request.Context())
		if err != nil {
			h.sendError(c, err)
			return
		}
	}

	if id != "" {
		profile, url, err = h.domain.GetProfileFromSubjectID(c.Request.Context(), id)
		if err != nil {
			h.sendError(c, err)
			return
		}
	}

	c.JSON(http.StatusOK, ProfileResponse{
		Alias:     profile.Alias,
		AvatarURL: url,
		Version:   profile.Version,
	})
}

func (h *Handler) GetProfiles(c *gin.Context) {
	alias := c.Param("alias")

	var req *GetProfilesRequest
	if err := c.BindJSON(&req); err != nil {
		h.sendError(c, err)
		return
	}

	next, profiles, urls, err := h.domain.GetProfilesFromAlias(c.Request.Context(), alias, req.Size, req.Page)
	if err != nil {
		h.sendError(c, err)
		return
	}

	res := make([]*ProfileResponse, len(profiles))
	for _, profile := range profiles {
		res = append(res, &ProfileResponse{
			Alias:     profile.Alias,
			AvatarURL: urls[profile.SubjectID],
			Version:   profile.Version,
		})
	}

	c.JSON(http.StatusOK, GetProfilesResponse{
		NextPage: next,
	})
}

func (h *Handler) AddProfile(c *gin.Context) {
	var req *AddProfileRequest
	if err := c.BindJSON(&req); err != nil {
		h.sendError(c, err)
		return
	}

	profile, url, err := h.domain.AddProfile(c.Request.Context(), req.Alias)
	if err != nil {
		h.sendError(c, err)
		return
	}

	c.JSON(http.StatusCreated, ProfileResponse{
		Alias:     profile.Alias,
		AvatarURL: url,
		Version:   profile.Version,
	})
}

func (h *Handler) UpdateProfileMetadata(c *gin.Context) {
	var req *UpdateProfileMetadataRequest
	if err := c.BindJSON(&req); err != nil {
		h.sendError(c, err)
		return
	}

	profile, url, err := h.domain.UpdateProfileMetadata(c.Request.Context(), req.Version, req.Alias)
	if err != nil {
		h.sendError(c, err)
		return
	}

	c.JSON(http.StatusOK, ProfileResponse{
		Alias:     profile.Alias,
		AvatarURL: url,
		Version:   profile.Version,
	})
}

func (h *Handler) UploadAvatar(c *gin.Context) {
	url, err := h.domain.UploadAvatar(c.Request.Context())
	if err != nil {
		h.sendError(c, err)
		return
	}

	c.JSON(http.StatusOK, UploadAvatarResponse{
		UploadURL: url,
	})
}

func (h *Handler) DeleteAvatar(c *gin.Context) {
	profile, url, err := h.domain.DeleteAvatar(c.Request.Context())
	if err != nil {
		h.sendError(c, err)
		return
	}

	c.JSON(http.StatusOK, ProfileResponse{
		Alias:     profile.Alias,
		AvatarURL: url,
		Version:   profile.Version,
	})
}

func (h *Handler) sendError(c *gin.Context, err error) {
	var code int

	if errors.Is(err, domain.ErrNotFound) {
		code = http.StatusNotFound
	}

	if code == 0 {
		code = http.StatusInternalServerError
	}

	c.Error(err)
	c.AbortWithStatus(code)
}

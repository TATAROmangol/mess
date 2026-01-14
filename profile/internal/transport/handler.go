package transport

import (
	"errors"
	"net/http"

	"github.com/TATAROmangol/mess/profile/internal/domain"
	"github.com/TATAROmangol/mess/profile/internal/model"
	"github.com/TATAROmangol/mess/profile/pkg/dto"
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

	c.JSON(http.StatusOK, dto.ProfileResponse{
		SubjectID: profile.SubjectID,
		Alias:     profile.Alias,
		AvatarURL: url,
		Version:   profile.Version,
	})
}

func (h *Handler) GetProfiles(c *gin.Context) {
	alias := c.Param("alias")

	var req *dto.GetProfilesRequest
	if err := c.BindJSON(&req); err != nil {
		h.sendError(c, err)
		return
	}

	next, profiles, urls, err := h.domain.GetProfilesFromAlias(c.Request.Context(), alias, req.Size, req.Page)
	if err != nil {
		h.sendError(c, err)
		return
	}

	res := make([]*dto.ProfileResponse, len(profiles))
	for _, profile := range profiles {
		res = append(res, &dto.ProfileResponse{
			SubjectID: profile.SubjectID,
			Alias:     profile.Alias,
			AvatarURL: urls[profile.SubjectID],
			Version:   profile.Version,
		})
	}

	c.JSON(http.StatusOK, dto.GetProfilesResponse{
		NextPage: next,
	})
}

func (h *Handler) AddProfile(c *gin.Context) {
	var req *dto.AddProfileRequest
	if err := c.BindJSON(&req); err != nil {
		h.sendError(c, err)
		return
	}

	profile, url, err := h.domain.AddProfile(c.Request.Context(), req.Alias)
	if err != nil {
		h.sendError(c, err)
		return
	}

	c.JSON(http.StatusCreated, dto.ProfileResponse{
		SubjectID: profile.SubjectID,
		Alias:     profile.Alias,
		AvatarURL: url,
		Version:   profile.Version,
	})
}

func (h *Handler) UpdateProfileMetadata(c *gin.Context) {
	var req *dto.UpdateProfileMetadataRequest
	if err := c.BindJSON(&req); err != nil {
		h.sendError(c, err)
		return
	}

	profile, url, err := h.domain.UpdateProfileMetadata(c.Request.Context(), req.Version, req.Alias)
	if err != nil {
		h.sendError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ProfileResponse{
		SubjectID: profile.SubjectID,
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

	c.JSON(http.StatusOK, dto.UploadAvatarResponse{
		UploadURL: url,
	})
}

func (h *Handler) DeleteAvatar(c *gin.Context) {
	profile, url, err := h.domain.DeleteAvatar(c.Request.Context())
	if err != nil {
		h.sendError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ProfileResponse{
		SubjectID: profile.SubjectID,
		Alias:     profile.Alias,
		AvatarURL: url,
		Version:   profile.Version,
	})
}

func (h *Handler) DeleteProfile(c *gin.Context) {
	profile, url, err := h.domain.DeleteProfile(c.Request.Context())
	if err != nil {
		h.sendError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ProfileResponse{
		SubjectID: profile.SubjectID,
		Alias:     profile.Alias,
		AvatarURL: url,
		Version:   profile.Version,
	})
}

func (h *Handler) sendError(c *gin.Context, err error) {
	var code int

	if errors.Is(err, domain.ErrNotFound) {
		code = http.StatusNoContent
	}

	if code == 0 {
		code = http.StatusInternalServerError
	}

	c.AbortWithError(code, err)
}

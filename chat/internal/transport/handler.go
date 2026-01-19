package transport

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/TATAROmangol/mess/chat/internal/ctxkey"
	"github.com/TATAROmangol/mess/chat/internal/domain"
	"github.com/TATAROmangol/mess/chat/pkg/dto"
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

func (h *Handler) AddChat(c *gin.Context) {
	secondSubjID := c.Param("subject_id")
	if secondSubjID == "" {
		h.sendError(c, InvalidRequestError)
		return
	}

	chat, err := h.domain.AddChat(c.Request.Context(), secondSubjID)
	if err != nil {
		h.sendError(c, err)
		return
	}

	lastReads, err := h.domain.GetLastReads(c.Request.Context(), chat.ID)
	if err != nil {
		h.sendError(c, err)
		return
	}
	lastReadsMap := map[string]int{}
	for _, lr := range lastReads {
		lastReadsMap[lr.SubjectID] = lr.MessageID
	}

	c.JSON(http.StatusCreated, dto.ChatResponse{
		ChatID:          chat.ID,
		SecondSubjectID: secondSubjID,
		LastReads:       lastReadsMap,
		Messages:        []*dto.MessageResponse{},
	})
}

func (h *Handler) GetChatBySubjectID(c *gin.Context) {
	id := c.Param("subject_id")
	if id == "" {
		h.sendError(c, InvalidRequestError)
		return
	}

	var limit int
	var err error

	sLimit := c.Query("limit")
	if sLimit != "" {
		limit, err = strconv.Atoi(sLimit)
		if err != nil {
			h.sendError(c, err)
			return
		}
	}

	chat, err := h.domain.GetChatBySubjectID(c.Request.Context(), id)
	if err != nil {
		h.sendError(c, err)
		return
	}

	h.returnChatResponse(c, chat.ID, limit)
}

func (h *Handler) GetChatByID(c *gin.Context) {
	var limit int
	var err error

	sLimit := c.Query("limit")
	if sLimit != "" {
		limit, err = strconv.Atoi(sLimit)
		if err != nil {
			h.sendError(c, err)
			return
		}
	}

	sChatID := c.Param("chat_id")
	if sChatID == "" {
		h.sendError(c, InvalidRequestError)
		return
	}

	chatID, err := strconv.Atoi(sChatID)
	if err != nil {
		h.sendError(c, err)
		return
	}

	h.returnChatResponse(c, chatID, limit)
}

func (h *Handler) returnChatResponse(c *gin.Context, chatID int, limit int) {
	subj, err := ctxkey.ExtractSubject(c.Request.Context())
	if err != nil {
		h.sendError(c, err)
		return
	}

	messages, err := h.domain.GetMessagesToLastRead(c.Request.Context(), chatID, limit)
	if err != nil {
		h.sendError(c, err)
		return
	}
	resMessages := MessagesModelToMessageDTO(messages)

	lastReads, err := h.domain.GetLastReads(c.Request.Context(), chatID)
	if err != nil {
		h.sendError(c, err)
		return
	}

	var secondID string
	lastReadsMap := map[string]int{}
	for _, lr := range lastReads {
		if lr.SubjectID != subj.GetSubjectId() {
			secondID = lr.SubjectID
		}
		lastReadsMap[lr.SubjectID] = lr.MessageID
	}

	if secondID == "" {
		h.sendError(c, fmt.Errorf("not found second subject id"))
		return
	}

	c.JSON(http.StatusOK, dto.ChatResponse{
		ChatID:          chatID,
		SecondSubjectID: secondID,
		LastReads:       lastReadsMap,
		Messages:        resMessages,
	})
}

func (h *Handler) GetChats(c *gin.Context) {
	sLimit := c.Query("limit")
	sBefore := c.Query("before")
	sAfter := c.Query("after")

	filter, err := MakeChatPaginationFilter(sLimit, sBefore, sAfter)
	if err != nil {
		h.sendError(c, err)
		return
	}

	chatsMetadata, err := h.domain.GetChatsMetadata(c.Request.Context(), filter)
	if err != nil {
		h.sendError(c, err)
		return
	}
	resChats := ChatsMetadataModelToDTO(chatsMetadata)

	c.JSON(http.StatusOK, resChats)
}

func (h *Handler) GetMessages(c *gin.Context) {
	sChat := c.Query("chat")
	sLimit := c.Query("limit")
	sBefore := c.Query("before")
	sAfter := c.Query("after")

	chatID, err := strconv.Atoi(sChat)
	if err != nil {
		return
	}

	filter, err := MakeMessagePaginationFilter(sLimit, sBefore, sAfter)
	if err != nil {
		h.sendError(c, err)
		return
	}

	messages, err := h.domain.GetMessages(c.Request.Context(), chatID, filter)
	if err != nil {
		h.sendError(c, err)
		return
	}
	resMessages := MessagesModelToMessageDTO(messages)

	c.JSON(http.StatusOK, resMessages)
}

func (h *Handler) AddMessage(c *gin.Context) {
	var req *dto.AddMessageRequest
	if err := c.BindJSON(&req); err != nil {
		h.sendError(c, err)
		return
	}

	mess, err := h.domain.SendMessage(c.Request.Context(), req.ChatID, req.Content)
	if err != nil {
		h.sendError(c, err)
		return
	}

	c.JSON(http.StatusCreated, MessageModelToMessageDTO(mess))
}

func (h *Handler) UpdateMessage(c *gin.Context) {
	var req *dto.UpdateMessageRequest
	if err := c.BindJSON(&req); err != nil {
		h.sendError(c, err)
		return
	}

	mess, err := h.domain.UpdateMessage(c.Request.Context(), req.MessageID, req.Content, req.Version)
	if err != nil {
		h.sendError(c, err)
		return
	}

	c.JSON(http.StatusCreated, MessageModelToMessageDTO(mess))
}

func (h *Handler) UpdateLastRead(c *gin.Context) {
	var req *dto.UpdateLastReadRequest
	if err := c.BindJSON(&req); err != nil {
		h.sendError(c, err)
		return
	}

	_, err := h.domain.UpdateLastRead(c.Request.Context(), req.ChatID, req.MessageID)
	if err != nil {
		h.sendError(c, err)
		return
	}

	c.Status(http.StatusOK)
}

func (h *Handler) sendError(c *gin.Context, err error) {
	var code int

	if errors.Is(err, InvalidRequestError) {
		code = http.StatusBadRequest
	}

	if errors.Is(err, domain.ErrNotFound) {
		code = http.StatusNoContent
	}

	if code == 0 {
		code = http.StatusInternalServerError
	}

	c.AbortWithError(code, err)
}

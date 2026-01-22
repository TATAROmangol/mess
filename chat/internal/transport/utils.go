package transport

import (
	"strconv"

	"github.com/TATAROmangol/mess/chat/internal/domain"
	"github.com/TATAROmangol/mess/chat/internal/model"
	httpdto "github.com/TATAROmangol/mess/shared/dto/http"
)

func MessageModelToMessageDTO(mess *model.Message) *httpdto.MessageResponse {
	return &httpdto.MessageResponse{
		ID:        mess.ID,
		Version:   mess.Version,
		Content:   mess.Content,
		SenderID:  mess.SenderSubjectID,
		CreatedAt: mess.CreatedAt,
	}
}

func MessagesModelToMessageDTO(messages []*model.Message) []*httpdto.MessageResponse {
	resMessages := make([]*httpdto.MessageResponse, 0, len(messages))
	for _, mess := range messages {
		resMessages = append(resMessages, MessageModelToMessageDTO(mess))
	}

	return resMessages
}

func ChatsMetadataModelToDTO(chatsMetadata []*model.ChatMetadata) []*httpdto.ChatsMetadataResponse {
	resChats := make([]*httpdto.ChatsMetadataResponse, 0, len(chatsMetadata))
	for _, cm := range chatsMetadata {
		resChats = append(resChats, &httpdto.ChatsMetadataResponse{
			ChatID:          cm.ChatID,
			SecondSubjectID: cm.SecondSubjectID,
			UpdatedAt:       cm.UpdatedAt,
			LastMessage: &httpdto.MessageResponse{
				ID:       cm.LastMessage.MessageID,
				Content:  cm.LastMessage.Content,
				SenderID: cm.LastMessage.SenderID,
			},
			UnreadCount:       cm.UnreadCount,
			IsLastMessageRead: cm.IsLastMessageRead,
		})
	}

	return resChats
}

func MakeMessagePaginationFilter(sLimit string, sBefore string, sAfter string) (*domain.MessagePaginationFilter, error) {
	if sBefore == "" && sAfter == "" || sAfter != "" && sBefore != "" {
		return nil, InvalidRequestError
	}

	filter := domain.MessagePaginationFilter{}

	var err error
	if sLimit != "" {
		filter.Limit, err = strconv.Atoi(sLimit)
		if err != nil {
			return nil, err
		}
	}

	var last int
	if sBefore != "" {
		last, err = strconv.Atoi(sBefore)
		if err != nil {
			return nil, err
		}
		filter.Direction = domain.DirectionBefore
	}
	if sAfter != "" {
		last, err = strconv.Atoi(sAfter)
		if err != nil {
			return nil, err
		}
		filter.Direction = domain.DirectionAfter
	}

	filter.LastMessageID = &last

	return &filter, nil
}

func MakeChatPaginationFilter(sLimit string, sBefore string, sAfter string) (*domain.ChatPaginationFilter, error) {
	if sAfter != "" && sBefore != "" {
		return nil, InvalidRequestError
	}

	filter := domain.ChatPaginationFilter{}

	var err error
	if sLimit != "" {
		filter.Limit, err = strconv.Atoi(sLimit)
		if err != nil {
			return nil, err
		}
	}

	var last int
	if sBefore != "" {
		last, err = strconv.Atoi(sBefore)
		if err != nil {
			return nil, err
		}
		filter.Direction = domain.DirectionBefore
	}
	if sAfter != "" {
		last, err = strconv.Atoi(sAfter)
		if err != nil {
			return nil, err
		}
		filter.Direction = domain.DirectionAfter
	}

	filter.LastChatID = &last

	return &filter, nil
}

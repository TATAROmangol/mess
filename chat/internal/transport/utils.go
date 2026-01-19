package transport

import (
	"strconv"

	"github.com/TATAROmangol/mess/chat/internal/domain"
	"github.com/TATAROmangol/mess/chat/internal/model"
	"github.com/TATAROmangol/mess/chat/pkg/dto"
)

func MessageModelToMessageDTO(mess *model.Message) *dto.MessageResponse {
	return &dto.MessageResponse{
		ID:        mess.ID,
		Version:   mess.Version,
		Content:   mess.Content,
		SenderID:  mess.SenderSubjectID,
		CreatedAt: mess.CreatedAt,
	}
}

func MessagesModelToMessageDTO(messages []*model.Message) []*dto.MessageResponse {
	resMessages := make([]*dto.MessageResponse, 0, len(messages))
	for _, mess := range messages {
		resMessages = append(resMessages, MessageModelToMessageDTO(mess))
	}

	return resMessages
}

func ChatsMetadataModelToDTO(chatsMetadata []*model.ChatMetadata) []*dto.ChatsMetadataResponse {
	resChats := make([]*dto.ChatsMetadataResponse, 0, len(chatsMetadata))
	for _, cm := range chatsMetadata {
		resChats = append(resChats, &dto.ChatsMetadataResponse{
			ChatID:          cm.ChatID,
			SecondSubjectID: cm.SecondSubjectID,
			UpdatedAt:       cm.UpdatedAt,
			LastMessage: &dto.MessageResponse{
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

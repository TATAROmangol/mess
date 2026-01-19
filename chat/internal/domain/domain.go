package domain

import (
	"context"
	"errors"
	"fmt"

	"github.com/TATAROmangol/mess/chat/internal/ctxkey"
	loglables "github.com/TATAROmangol/mess/chat/internal/loglables"
	"github.com/TATAROmangol/mess/chat/internal/model"
	"github.com/TATAROmangol/mess/chat/internal/storage"

	"github.com/TATAROmangol/mess/shared/utils"
)

func (d *Domain) AddChat(ctx context.Context, secondSubjectID string) (*model.Chat, error) {
	subj, err := ctxkey.ExtractSubject(ctx)
	if err != nil {
		return nil, fmt.Errorf("extract subject: %w", err)
	}
	lg, err := ctxkey.ExtractLogger(ctx)
	if err != nil {
		return nil, fmt.Errorf("extract logger: %w", err)
	}

	tx, err := d.Storage.WithTransaction(ctx)
	if err != nil {
		return nil, fmt.Errorf("storage with transaction: %w", err)
	}
	defer tx.Rollback()

	chat, err := tx.Chat().CreateChat(ctx, subj.GetSubjectId(), secondSubjectID)
	if err != nil {
		return nil, fmt.Errorf("create chat: %w", err)
	}
	lg = lg.With(loglables.Chat, *chat)

	lastReadSubj, err := tx.LastRead().CreateLastRead(ctx, subj.GetSubjectId(), chat.ID)
	if err != nil {
		return nil, fmt.Errorf("create last read subj: %w", err)
	}
	lg = lg.With(loglables.LastReadSubject, *lastReadSubj)

	lastReadSecond, err := tx.LastRead().CreateLastRead(ctx, secondSubjectID, chat.ID)
	if err != nil {
		return nil, fmt.Errorf("create last read second subj: %w", err)
	}
	lg = lg.With(loglables.LastReadSecond, *lastReadSecond)

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}
	lg.Debug("add last reads with chat")

	return chat, nil
}

type LastReadsPair struct {
	Mine  *model.LastRead
	Other *model.LastRead
}

func (d *Domain) GetChats(ctx context.Context, filter *ChatPaginationFilter) ([]*model.ChatMetadata, error) {
	subj, err := ctxkey.ExtractSubject(ctx)
	if err != nil {
		return nil, fmt.Errorf("extract subject: %w", err)
	}

	storageFilter := DefaultPaginationChat
	switch filter.Direction {
	case DirectionBefore:
		storageFilter.Asc = false
	default:
		storageFilter.Asc = true
	}

	storageFilter.LastID = filter.LastChatID

	chats, err := d.Storage.Chat().GetChatsBySubjectID(ctx, subj.GetSubjectId(), &storageFilter)
	if err != nil {
		return nil, fmt.Errorf("get chats bu subject id: %w", err)
	}

	if len(chats) == 0 {
		return []*model.ChatMetadata{}, nil
	}

	lastReads, err := d.Storage.LastRead().GetLastReadsByChatIDs(ctx, model.GetChatsID(chats))
	if err != nil {
		return nil, fmt.Errorf("get last read by chat ids: %w", err)
	}

	lastReadsMap := map[int]*LastReadsPair{}
	for _, read := range lastReads {
		pair, ok := lastReadsMap[read.ChatID]
		if !ok {
			pair = &LastReadsPair{}
			lastReadsMap[read.ChatID] = pair
		}

		if read.SubjectID == subj.GetSubjectId() {
			pair.Mine = read
		} else {
			pair.Other = read
		}
	}

	lastMessages, err := d.Storage.Message().GetLastMessagesByChatsID(ctx, model.GetChatsID(chats))
	if err != nil {
		return nil, fmt.Errorf("get last messages by chats id: %w", err)
	}

	lastMessagesMap := map[int]*model.Message{}
	for _, mes := range lastMessages {
		lastMessagesMap[mes.ChatID] = mes
	}

	res := make([]*model.ChatMetadata, 0, len(chats))
	for _, chat := range chats {
		meta := &model.ChatMetadata{
			ChatID: chat.ID,
		}

		mes, ok := lastMessagesMap[chat.ID]
		if !ok {
			res = append(res, meta)
			continue
		}

		meta.LastMessage = &model.LastMessage{
			MessageID: mes.ID,
			Content:   mes.Content,
			SenderID:  mes.SenderSubjectID,
		}

		pair := lastReadsMap[chat.ID]
		if mes.SenderSubjectID == subj.GetSubjectId() {
			meta.UnreadCount = 0
			meta.IsLastMessageRead = pair.Other.MessageNumber >= mes.Number
			res = append(res, meta)
			continue
		}

		meta.UnreadCount = chat.MessagesCount - pair.Mine.MessageNumber
		meta.IsLastMessageRead = true

		res = append(res, meta)
	}

	return res, nil
}

func (d *Domain) GetLastReads(ctx context.Context, chatID int) ([]*model.LastRead, error) {
	subj, err := ctxkey.ExtractSubject(ctx)
	if err != nil {
		return nil, fmt.Errorf("extract subject: %w", err)
	}

	lastReads, err := d.Storage.LastRead().GetLastReadsByChatID(ctx, chatID)
	if err != nil {
		return nil, fmt.Errorf("get last read by chat id: %w", err)
	}

	if lastReads[0].SubjectID != subj.GetSubjectId() && lastReads[1].SubjectID != subj.GetSubjectId() {
		return nil, SubjectNotHaveThisResource
	}

	return lastReads, err
}

func (d *Domain) UpdateLastRead(ctx context.Context, chatID int, messageID int, messageNumber int) (*model.LastRead, error) {
	subj, err := ctxkey.ExtractSubject(ctx)
	if err != nil {
		return nil, fmt.Errorf("extract subject: %w", err)
	}

	lastRead, err := d.Storage.LastRead().UpdateLastRead(ctx, subj.GetSubjectId(), chatID, messageID, messageNumber)
	if err != nil {
		return nil, fmt.Errorf("update last read: %w", err)
	}

	return lastRead, nil
}

func (d *Domain) GetMessages(ctx context.Context, chatID int, filter *MessagePaginationFilter) ([]*model.Message, error) {
	subj, err := ctxkey.ExtractSubject(ctx)
	if err != nil {
		return nil, fmt.Errorf("extract subject: %w", err)
	}
	chat, err := d.Storage.Chat().GetChatByID(ctx, chatID)
	if err != nil {
		return nil, fmt.Errorf("get chat by in: %w", err)
	}
	if chat.FirstSubjectID != subj.GetSubjectId() && chat.SecondSubjectID != subj.GetSubjectId() {
		return nil, SubjectNotHaveThisResource
	}

	storageFilter := DefaultPaginationMessage
	switch filter.Direction {
	case DirectionBefore:
		storageFilter.Asc = false
	default:
		storageFilter.Asc = true
	}

	storageFilter.LastID = filter.LastMessageID

	messages, err := d.Storage.Message().GetMessagesByChatID(ctx, chatID, &storageFilter)
	if err != nil {
		return nil, fmt.Errorf("get messages by chat id: %w", err)
	}

	if len(messages) == 0 {
		return messages, nil
	}

	if !storageFilter.Asc {
		utils.ReverseSlice(messages)
	}

	return messages, nil
}

func (d *Domain) GetMessagesToLastRead(ctx context.Context, chatID int) ([]*model.Message, error) {
	subj, err := ctxkey.ExtractSubject(ctx)
	if err != nil {
		return nil, fmt.Errorf("extract subject: %w", err)
	}
	lg, err := ctxkey.ExtractLogger(ctx)
	if err != nil {
		return nil, fmt.Errorf("extract logger: %w", err)
	}

	chat, err := d.Storage.Chat().GetChatByID(ctx, chatID)
	if err != nil {
		return nil, fmt.Errorf("get chat from id: %w", err)
	}
	lg = lg.With(loglables.Chat, *chat)

	lastRead, err := d.Storage.LastRead().GetLastReadBySubjectID(ctx, subj.GetSubjectId(), chatID)
	if err != nil {
		return nil, fmt.Errorf("get last read by chat id: %w", err)
	}
	lg = lg.With(loglables.LastRead, *lastRead)

	filter := DefaultPaginationMessage
	if chat.MessagesCount-lastRead.MessageNumber > filter.Limit {
		filter.LastID = &lastRead.MessageID
		filter.Asc = true
	}

	messages, err := d.Storage.Message().GetMessagesByChatID(ctx, chatID, &filter)
	if err != nil {
		return nil, fmt.Errorf("get messages by chat id: %w", err)
	}

	if len(messages) == 0 {
		return messages, nil
	}

	if !filter.Asc {
		utils.ReverseSlice(messages)
	}

	lastMess := messages[len(messages)-1]
	lastRead, err = d.Storage.LastRead().UpdateLastRead(ctx, subj.GetSubjectId(), chatID, lastMess.ID, lastMess.Number)
	if err != nil && !errors.Is(err, storage.ErrNoRows) {
		return nil, fmt.Errorf("update last read: %w", err)
	}
	if lastRead != nil {
		lg = lg.With(loglables.Updated, *lastRead)
	}

	lg.Debug("get messages to last read")

	return messages, nil
}

func (d *Domain) SendMessage(ctx context.Context, chatID int, content string) (*model.Message, error) {
	subj, err := ctxkey.ExtractSubject(ctx)
	if err != nil {
		return nil, fmt.Errorf("extract subject: %w", err)
	}
	lg, err := ctxkey.ExtractLogger(ctx)
	if err != nil {
		return nil, fmt.Errorf("extract logger: %w", err)
	}

	tx, err := d.Storage.WithTransaction(ctx)
	if err != nil {
		return nil, fmt.Errorf("storage with transaction: %w", err)
	}
	defer tx.Rollback()

	chat, err := tx.Chat().IncrementChatMessageNumber(ctx, chatID)
	if err != nil {
		return nil, fmt.Errorf("increment chat message number: %w", err)
	}
	lg = lg.With(loglables.Chat, *chat)

	message, err := tx.Message().CreateMessage(ctx, chatID, subj.GetSubjectId(), content, chat.MessagesCount)
	if err != nil {
		return nil, fmt.Errorf("create message: %w", err)
	}
	lg = lg.With(loglables.Message, *message)

	lastRead, err := tx.LastRead().UpdateLastRead(ctx, subj.GetSubjectId(), chatID, message.ID, message.Number)
	if err != nil {
		return nil, fmt.Errorf("update last read: %w", err)
	}
	lg = lg.With(loglables.LastRead, *lastRead)

	outbox, err := tx.MessageOutbox().AddMessageOutbox(ctx, chatID, message.ID, model.AddOperation)
	if err != nil {
		return nil, fmt.Errorf("add message outbox: %w", err)
	}
	lg = lg.With(loglables.MessageOutbox, *outbox)

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}
	lg.Debug("send message")

	return message, nil
}

func (d *Domain) UpdateMessage(ctx context.Context, messageID int, content string, version int) (*model.Message, error) {
	subj, err := ctxkey.ExtractSubject(ctx)
	if err != nil {
		return nil, fmt.Errorf("extract subject: %w", err)
	}
	mess, err := d.Storage.Message().GetMessageByID(ctx, messageID)
	if err != nil {
		return nil, fmt.Errorf("get chat by in: %w", err)
	}
	if mess.SenderSubjectID != subj.GetSubjectId() {
		return nil, SubjectNotHaveThisResource
	}

	lg, err := ctxkey.ExtractLogger(ctx)
	if err != nil {
		return nil, fmt.Errorf("extract logger: %w", err)
	}

	tx, err := d.Storage.WithTransaction(ctx)
	if err != nil {
		return nil, fmt.Errorf("storage with transaction: %w", err)
	}
	defer tx.Rollback()

	message, err := tx.Message().UpdateMessageContent(ctx, messageID, content, version)
	if err != nil {
		return nil, fmt.Errorf("update message content: %w", err)
	}
	lg = lg.With(loglables.Message, *message)

	chat, err := tx.Chat().GetChatByID(ctx, message.ChatID)
	if err != nil {
		return nil, fmt.Errorf("get chat by chat id: %w", err)
	}
	lg = lg.With(loglables.Chat, *chat)

	outbox, err := tx.MessageOutbox().AddMessageOutbox(ctx, chat.ID, message.ID, model.UpdateOperation)
	if err != nil {
		return nil, fmt.Errorf("add message outbox: %w", err)
	}
	lg = lg.With(loglables.MessageOutbox, *outbox)

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}
	lg.Debug("update message")

	return message, nil
}

package central

import (
	"context"
	"fmt"
	"math"
	"seisami/server/centraldb"
	"seisami/server/types"
	"seisami/server/utils"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type NotificationService struct {
	pool    *pgxpool.Pool
	queries *centraldb.Queries
}

type NotificationsPaginatedResponse struct {
	Notifications []types.Notification `json:"notifications"`
	TotalCount    int64                `json:"total_count"`
	TotalPages    int                  `json:"total_pages"`
	CurrentPage   int                  `json:"current_page"`
	PageSize      int                  `json:"page_size"`
}

func NewNotificationService(pool *pgxpool.Pool, queries *centraldb.Queries) *NotificationService {
	return &NotificationService{pool, queries}
}

func (s *NotificationService) createNotification(ctx context.Context, userId uuid.UUID, title, message, nType, target string) error {
	notifID := uuid.New()
	_, err := s.queries.CreateNotification(ctx, centraldb.CreateNotificationParams{
		ID: pgtype.UUID{
			Bytes: notifID,
			Valid: true,
		},
		UserID:  pgtype.UUID{Bytes: userId, Valid: true},
		Title:   title,
		Message: message,
		Type:    nType,
		Target:  pgtype.Text{String: target, Valid: target != ""},
		Read: pgtype.Bool{
			Bool:  false,
			Valid: true,
		},
	})
	return err
}

func (s *NotificationService) markAsRead(ctx context.Context, notificationID, userId uuid.UUID) error {
	fmt.Println("marking notification read::")
	return s.queries.MarkNotificationAsRead(ctx, centraldb.MarkNotificationAsReadParams{
		ID: pgtype.UUID{
			Bytes: notificationID,
			Valid: true,
		},
		UserID: pgtype.UUID{
			Bytes: userId,
			Valid: true,
		},
	})
}

func (s *NotificationService) getNotifications(ctx context.Context, userUUID uuid.UUID, offset int32) (*NotificationsPaginatedResponse, error) {
	const pageSize = 20

	totalCount, err := s.queries.CountNotificationsForUser(ctx, pgtype.UUID{
		Bytes: userUUID,
		Valid: true,
	})
	if err != nil {
		return nil, err
	}

	notifications, err := s.queries.GetNotificationsForUser(ctx, centraldb.GetNotificationsForUserParams{
		UserID: pgtype.UUID{
			Bytes: userUUID,
			Valid: true,
		},
		Offset: offset,
		Limit:  pageSize,
	})
	if err != nil {
		return nil, err
	}

	totalPages := int(math.Ceil(float64(totalCount) / float64(pageSize)))
	currentPage := int(offset/pageSize) + 1

	notifs := make([]types.Notification, 0)

	for _, n := range notifications {
		notifs = append(notifs, types.Notification{
			Title:   n.Title,
			Message: n.Message,
			Type:    n.Type,
			Target:  n.Target.String,
			ID:      n.ID.String(),
			Read:    n.Read.Bool,
			Time:    utils.ConvertTimestamptzToLocal(n.CreatedAt),
		})
	}

	return &NotificationsPaginatedResponse{
		Notifications: notifs,
		TotalCount:    totalCount,
		TotalPages:    totalPages,
		CurrentPage:   currentPage,
		PageSize:      pageSize,
	}, nil
}

package repository

import (
	"context"
	"testing"
	"time"

	db "github.com/Rain-kl/Wavelet/internal/infra/persistence"
	"github.com/Rain-kl/Wavelet/internal/model"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestAddUserTrafficUsageAccumulatesByUserAndMonth(t *testing.T) {
	conn, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, conn.AutoMigrate(&model.UserTrafficMonthly{}))
	db.SetDB(conn)
	t.Cleanup(func() { db.SetDB(nil) })

	loggedAt := time.Date(2026, time.August, 4, 12, 0, 0, 0, time.UTC)
	records := []*model.OpenFlareAccessLog{
		{OwnerID: 7, LoggedAt: loggedAt, BytesSent: 100},
		{OwnerID: 7, LoggedAt: loggedAt.Add(2 * time.Hour), BytesSent: 50},
		{OwnerID: 0, LoggedAt: loggedAt, BytesSent: 1000},
	}
	require.NoError(t, AddUserTrafficUsage(context.Background(), records))
	require.NoError(t, AddUserTrafficUsage(context.Background(), records[:1]))

	usage, err := ListUserTrafficUsage(context.Background(), time.Date(2026, time.August, 1, 0, 0, 0, 0, time.UTC))
	require.NoError(t, err)
	require.Equal(t, int64(250), usage[7])
}

func TestReserveUserPublishIsAtomicAtLimit(t *testing.T) {
	conn, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, conn.AutoMigrate(&model.UserPublishDailyCounter{}))
	db.SetDB(conn)
	t.Cleanup(func() { db.SetDB(nil) })

	day := time.Date(2026, time.August, 4, 0, 0, 0, 0, time.UTC)
	reserved, err := ReserveUserPublish(context.Background(), 9, day, 1)
	require.NoError(t, err)
	require.True(t, reserved)
	reserved, err = ReserveUserPublish(context.Background(), 9, day, 1)
	require.NoError(t, err)
	require.False(t, reserved)
	require.NoError(t, ReleaseUserPublish(context.Background(), 9, day))
	reserved, err = ReserveUserPublish(context.Background(), 9, day, 1)
	require.NoError(t, err)
	require.True(t, reserved)
}

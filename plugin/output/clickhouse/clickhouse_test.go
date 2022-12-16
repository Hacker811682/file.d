package clickhouse

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"
	"fmt"

	"github.com/golang/mock/gomock"
	"github.com/ozontech/file.d/logger"
	"github.com/ozontech/file.d/metric"
	"github.com/ozontech/file.d/pipeline"
	mock_ch "github.com/ozontech/file.d/plugin/output/clickhouse/mock"
	"github.com/stretchr/testify/require"
	insaneJSON "github.com/vitkovskii/insane-json"
)

func TestPrivateOut(t *testing.T) {
	testLogger := logger.Instance

	root := insaneJSON.Spawn()
	defer insaneJSON.Release(root)

	columns := []ConfigColumn{
		{
			Name:       "str_uni_1",
			ColumnType: "string",
		},
		{
			Name:       "int_uni_1",
			ColumnType: "int",
		},
		{
			Name:       "int_1",
			ColumnType: "int",
		},
		{
			Name:       "timestamp_1",
			ColumnType: "timestamp",
		},
	}

	strUniValue := "str_uni_1_value"
	intUniValue := 11
	intValue := 10
	timestampValue := 100

	root.AddField(columns[0].Name).MutateToString(strUniValue)
	root.AddField(columns[1].Name).MutateToInt(intUniValue)
	root.AddField(columns[2].Name).MutateToInt(intValue)
	root.AddField(columns[3].Name).MutateToInt(timestampValue)

	table := "table1"

	config := Config{
		Columns: columns,
		Retry:   3,
	}

	ctl := gomock.NewController(t)
	defer ctl.Finish()
	mockdb := mock_ch.NewMockDBIface(ctl)
	db := mockdb

	ctx := context.Background()
	var ctxMock = reflect.TypeOf((*context.Context)(nil)).Elem()

	mockdb.EXPECT().ExecContext(
		gomock.AssignableToTypeOf(ctxMock),
		"INSERT INTO table1 (str_uni_1,int_uni_1,int_1,timestamp_1) VALUES (?,?,?,?)",
		[]any{strUniValue, intUniValue, intValue, time.Unix(int64(timestampValue), 0).Format(time.RFC3339)},
	).Return(&resultForTest{}, nil).Times(1)

	builder, err := NewQueryBuilder(columns, table)
	require.NoError(t, err)

	p := &Plugin{
		config:       &config,
		queryBuilder: builder,
		conn:         db,
		logger:       testLogger,
		ctx:          ctx,
	}

	p.RegisterMetrics(metric.New("test"))

	batch := &pipeline.Batch{Events: []*pipeline.Event{{Root: root}}}
	p.out(nil, batch)
}

func TestPrivateOutWithRetry(t *testing.T) {
	testLogger := logger.Instance

	root := insaneJSON.Spawn()
	defer insaneJSON.Release(root)

	columns := []ConfigColumn{
		{
			Name:       "str_uni_1",
			ColumnType: "string",
		},
		{
			Name:       "int_1",
			ColumnType: "int",
		},
		{
			Name:       "timestamp_1",
			ColumnType: "timestamp",
		},
	}

	strUniValue := "str_uni_1_value"
	intValue := 10
	timestampValue := 100

	root.AddField(columns[0].Name).MutateToString(strUniValue)
	root.AddField(columns[1].Name).MutateToInt(intValue)
	root.AddField(columns[2].Name).MutateToInt(timestampValue)

	table := "table1"

	config := Config{
		Columns: columns,
		Retry:   3,
	}

	ctl := gomock.NewController(t)
	defer ctl.Finish()
	mockdb := mock_ch.NewMockDBIface(ctl)
	db := mockdb

	ctx := context.Background()
	var ctxMock = reflect.TypeOf((*context.Context)(nil)).Elem()

	mockdb.EXPECT().ExecContext(
		gomock.AssignableToTypeOf(ctxMock),
		"INSERT INTO table1 (str_uni_1,int_1,timestamp_1) VALUES (?,?,?)",
		[]any{strUniValue, intValue, time.Unix(int64(timestampValue), 0).Format(time.RFC3339)},
	).Return(&resultForTest{}, errors.New("someError")).Times(2)
	mockdb.EXPECT().ExecContext(
		gomock.AssignableToTypeOf(ctxMock),
		"INSERT INTO table1 (str_uni_1,int_1,timestamp_1) VALUES (?,?,?)",
		[]any{strUniValue, intValue, time.Unix(int64(timestampValue), 0).Format(time.RFC3339)},
	).Return(&resultForTest{}, nil).Times(1)

	builder, err := NewQueryBuilder(columns, table)
	require.NoError(t, err)

	p := &Plugin{
		config:       &config,
		queryBuilder: builder,
		conn:         db,
		logger:       testLogger,
		ctx:          ctx,
	}

	p.RegisterMetrics(metric.New("test"))

	batch := &pipeline.Batch{Events: []*pipeline.Event{{Root: root}}}
	p.out(nil, batch)
}

func TestPrivateOutNoGoodEvents(t *testing.T) {
	testLogger := logger.Instance

	root := insaneJSON.Spawn()
	defer insaneJSON.Release(root)

	columns := []ConfigColumn{
		{
			Name:       "str_uni_1",
			ColumnType: "string",
		},
		{
			Name:       "int_1",
			ColumnType: "int",
		},
		{
			Name:       "timestamp_1",
			ColumnType: "timestamp",
		},
	}

	strUniValue := "str_uni_1_value"
	intValue := 10

	// timestamp valur wasn't sent.
	root.AddField(columns[0].Name).MutateToString(strUniValue)
	root.AddField(columns[1].Name).MutateToInt(intValue)

	table := "table1"

	config := Config{
		Columns: columns,
		Retry:   3,
	}

	builder, err := NewQueryBuilder(columns, table)
	require.NoError(t, err)
	
	p := &Plugin{
		config:       &config,
		queryBuilder: builder,
		logger:       testLogger,
	}
	
	p.RegisterMetrics(metric.New("test"))
	
	batch := &pipeline.Batch{Events: []*pipeline.Event{{Root: root}}}
	fmt.Println(1)
	p.out(nil, batch)
	fmt.Println(2)
}

func TestPrivateOutWrongTypeInField(t *testing.T) {
	testLogger := logger.Instance

	root := insaneJSON.Spawn()
	defer insaneJSON.Release(root)

	columns := []ConfigColumn{
		{
			Name:       "str_uni_1",
			ColumnType: "string",
		},
		{
			Name:       "int_uni_1",
			ColumnType: "int",
		},
		{
			Name:       "int_1",
			ColumnType: "int",
		},
		{
			Name:       "timestamp_1",
			ColumnType: "timestamp",
		},
	}

	strUniValue := "str_uni_1_value"
	intUniValue := 11
	intValue := 10
	timestampValue := "100"

	root.AddField(columns[0].Name).MutateToString(strUniValue)
	root.AddField(columns[1].Name).MutateToInt(intUniValue)
	root.AddField(columns[2].Name).MutateToInt(intValue)
	// instead of 100 sender put "100" to json. Message'll be truncated.
	root.AddField(columns[3].Name).MutateToString(timestampValue)

	table := "table1"

	config := Config{Columns: columns, Retry: 3}

	ctl := gomock.NewController(t)
	defer ctl.Finish()

	builder, err := NewQueryBuilder(columns, table)
	require.NoError(t, err)

	p := &Plugin{
		config:       &config,
		queryBuilder: builder,
		logger:       testLogger,
	}

	p.RegisterMetrics(metric.New("test"))

	batch := &pipeline.Batch{Events: []*pipeline.Event{{Root: root}}}
	p.out(nil, batch)
}

func TestPrivateOutFewUniqueEventsYetWithDeduplicationEventsAndbadEvents(t *testing.T) {
	testLogger := logger.Instance

	root := insaneJSON.Spawn()
	defer insaneJSON.Release(root)

	secondUniqueRoot := insaneJSON.Spawn()
	defer insaneJSON.Release(secondUniqueRoot)

	badRoot := insaneJSON.Spawn()
	defer insaneJSON.Release(badRoot)

	columns := []ConfigColumn{
		{
			Name:       "str_uni_1",
			ColumnType: "string",
		},
		{
			Name:       "int_uni_1",
			ColumnType: "int",
		},
		{
			Name:       "int_1",
			ColumnType: "int",
		},
		{
			Name:       "timestamp_1",
			ColumnType: "timestamp",
		},
	}

	strUniValue := "str_uni_1_value"
	intUniValue := 11
	intValue := 10
	timestampValue := 100

	secStrUniValue := "str_uni_1_value____"
	secIntUniValue := 11000
	secIntValue := 10999
	secTimestampValue := 1008

	badTimestampValue := "100"

	root.AddField(columns[0].Name).MutateToString(strUniValue)
	root.AddField(columns[1].Name).MutateToInt(intUniValue)
	root.AddField(columns[2].Name).MutateToInt(intValue)
	root.AddField(columns[3].Name).MutateToInt(timestampValue)

	secondUniqueRoot.AddField(columns[0].Name).MutateToString(secStrUniValue)
	secondUniqueRoot.AddField(columns[1].Name).MutateToInt(secIntUniValue)
	secondUniqueRoot.AddField(columns[2].Name).MutateToInt(secIntValue)
	secondUniqueRoot.AddField(columns[3].Name).MutateToInt(secTimestampValue)

	badRoot.AddField(columns[0].Name).MutateToString(strUniValue)
	badRoot.AddField(columns[1].Name).MutateToInt(intUniValue)
	badRoot.AddField(columns[2].Name).MutateToInt(intValue)
	// instead of 100 sender put "100" to json. Message'll be truncated.
	badRoot.AddField(columns[3].Name).MutateToString(badTimestampValue)

	// This duplications will be removed from final query.
	rootDuplication := root
	rootDuplicationMore := root

	table := "table1"

	config := Config{
		Columns: columns,
		Retry:   3,
	}

	ctl := gomock.NewController(t)
	defer ctl.Finish()
	mockdb := mock_ch.NewMockDBIface(ctl)
	db := mockdb

	ctx := context.Background()
	var ctxMock = reflect.TypeOf((*context.Context)(nil)).Elem()

	mockdb.EXPECT().ExecContext(
		gomock.AssignableToTypeOf(ctxMock),
		"INSERT INTO table1 (str_uni_1,int_uni_1,int_1,timestamp_1) VALUES ($1,$2,$3,$4),($5,$6,$7,$8) ON CONFLICT(str_uni_1,int_uni_1) DO UPDATE SET int_1=EXCLUDED.int_1,timestamp_1=EXCLUDED.timestamp_1",
		[]any{strUniValue, intUniValue, intValue, time.Unix(int64(timestampValue), 0).Format(time.RFC3339),
			secStrUniValue, secIntUniValue, secIntValue, time.Unix(int64(secTimestampValue), 0).Format(time.RFC3339)},
	).Return(&resultForTest{}, nil).Times(1)

	builder, err := NewQueryBuilder(columns, table)
	require.NoError(t, err)

	p := &Plugin{
		config:       &config,
		queryBuilder: builder,
		conn:         db,
		logger:       testLogger,
		ctx:          ctx,
	}

	p.RegisterMetrics(metric.New("test"))

	batch := &pipeline.Batch{Events: []*pipeline.Event{
		{Root: root},
		{Root: rootDuplication},
		{Root: rootDuplicationMore},
		{Root: secondUniqueRoot},
		{Root: badRoot},
	}}
	p.out(nil, batch)
}

// TODO replace with gomock
type resultForTest struct{}

func (r resultForTest) LastInsertId() (int64, error)                    { return 1, nil }
func (r resultForTest) RowsAffected() (int64, error)                    { return 1, nil }

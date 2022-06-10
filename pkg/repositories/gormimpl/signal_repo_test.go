package gormimpl

import (
	"context"
	"reflect"
	"testing"

	"github.com/flyteorg/flyteadmin/pkg/repositories/errors"
	"github.com/flyteorg/flyteadmin/pkg/repositories/models"

	mockScope "github.com/flyteorg/flytestdlib/promutils"

	mocket "github.com/Selvatico/go-mocket"

	"github.com/stretchr/testify/assert"
)

var (
	signalModel = &models.Signal{
		SignalKey: models.SignalKey{
			ExecutionKey: models.ExecutionKey{
				Project: "project",
				Domain:  "domain",
				Name:    "name",
			},
			SignalID: "signal",
		},
		Type:  []byte{1, 2},
		Value: []byte{3, 4},
	}
)

func toSignalMap(signalModel models.Signal) map[string]interface{} {
	signal := make(map[string]interface{})
	signal["execution_project"] = signalModel.Project
	signal["execution_domain"] = signalModel.Domain
	signal["execution_name"] = signalModel.Name
	signal["signal_id"] = signalModel.SignalID
	if signalModel.Type != nil {
		signal["type"] = signalModel.Type
	}
	if signalModel.Value != nil {
		signal["value"] = signalModel.Value
	}

	return signal
}

func TestGetOrCreateSignal(t *testing.T) {
	ctx := context.Background()

	signalRepo := NewSignalRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())

	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.Logging = true

	// create initial signalModel
	mockInsertQuery := GlobalMock.NewMock()
	mockInsertQuery.WithQuery(
		`INSERT INTO "signals" ("id","created_at","updated_at","deleted_at","execution_project","execution_domain","execution_name","signal_id","type","value") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`)

	err := signalRepo.GetOrCreate(ctx, signalModel)
	assert.NoError(t, err)

	assert.True(t, mockInsertQuery.Triggered)
	mockInsertQuery.Triggered = false // reset to false for second call

	// initialize query mocks
	signalModels := []map[string]interface{}{toSignalMap(*signalModel)}
	mockSelectQuery := GlobalMock.NewMock()
	mockSelectQuery.WithQuery(
		`SELECT * FROM "signals" WHERE "signals"."created_at" = $1 AND "signals"."updated_at" = $2 AND "signals"."execution_project" = $3 AND "signals"."execution_domain" = $4 AND "signals"."execution_name" = $5 AND "signals"."signal_id" = $6 AND "signals"."execution_project" = $7 AND "signals"."execution_domain" = $8 AND "signals"."execution_name" = $9 AND "signals"."signal_id" = $10 ORDER BY "signals"."id" LIMIT 1`).WithReply(signalModels)

	// retrieve existing signalModel
	lookupSignalModel := &models.Signal{}
	*lookupSignalModel = *signalModel
	lookupSignalModel.Type = nil
	lookupSignalModel.Value = nil

	err = signalRepo.GetOrCreate(ctx, lookupSignalModel)
	assert.NoError(t, err)
	assert.True(t, reflect.DeepEqual(signalModel, lookupSignalModel))

	assert.True(t, mockSelectQuery.Triggered)
	assert.False(t, mockInsertQuery.Triggered)
}

// TODO - fix list
/*func TestListSignals(t *testing.T) {
	ctx := context.Background()

	signalRepo := NewSignalRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())

	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.Logging = true

	// list signalModels
	signalModels := []map[string]interface{}{toSignalMap(*signalModel)}
	mockSelectQuery := GlobalMock.NewMock()
	mockSelectQuery.WithQuery(
		//`SELECT * FROM "signals" WHERE "signals"."execution_project" = $1 AND "signals"."execution_domain" = $2 AND "signals"."execution_name" = $3 AND "signals"."signal_id" = $4 AND "signals"."type" = $5 AND "signals"."value" = $6`).WithReply(signalModels)
		`SELECT * FROM "signals" WHERE "signals"."created_at" = $1 AND "signals"."updated_at" = $2 AND "signals"."execution_project" = $3 AND "signals"."execution_domain" = $4 AND "signals"."execution_name" = $5 AND "signals"."signal_id" = $6 AND "signals"."type" = $7 AND "signals"."value" = $8`).WithReply(signalModels)

	signals, err := signalRepo.List(ctx, *signalModel)
	assert.NoError(t, err)

	assert.True(t, reflect.DeepEqual([]models.Signal{*signalModel}, signals))
	assert.True(t, mockSelectQuery.Triggered)
}*/

func TestUpdateSignal(t *testing.T) {
	ctx := context.Background()

	signalRepo := NewSignalRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())

	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.Logging = true

	// update signalModel does not exits
	mockUpdateQuery := GlobalMock.NewMock()
	mockUpdateQuery.WithQuery(
		`UPDATE "signals" SET "updated_at"=$1,"value"=$2 WHERE "execution_project" = $3 AND "execution_domain" = $4 AND "execution_name" = $5 AND "signal_id" = $6`).WithRowsNum(0)

	err := signalRepo.Update(ctx, *signalModel)
	assert.Error(t, err)

	assert.True(t, mockUpdateQuery.Triggered)
	mockUpdateQuery.Triggered = false // reset to false for second call

	// update signalModel exists
	mockUpdateQuery.WithRowsNum(1)

	err = signalRepo.Update(ctx, *signalModel)
	assert.NoError(t, err)

	assert.True(t, mockUpdateQuery.Triggered)
}

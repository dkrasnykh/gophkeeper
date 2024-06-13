package storage

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/dkrasnykh/gophkeeper/pkg/models"
)

var (
	text1 = models.Text{Type: models.TextItem, Tag: "tag1", Key: "key1", Value: "value 1", Comment: "comment", Created: 1}
	text2 = models.Text{Type: models.TextItem, Tag: "tag1", Key: "KEY2", Value: "value 2", Comment: "comment 2", Created: 2}
)

type TextStorager interface {
	All(ctx context.Context) ([]models.Text, error)
	ByKey(ctx context.Context, key string) (models.Text, error)
	Save(ctx context.Context, text models.Text) error
	Update(ctx context.Context, text models.Text) error
}

type testTextStorager interface {
	TextStorager
	clean(ctx context.Context) error
}

type TextSqliteTestSuite struct {
	suite.Suite
	testTextStorager
}

func (ts *TextSqliteTestSuite) SetupSuite() {
	_ = Migrate("client_test.db")
	ts.testTextStorager, _ = NewTextSqlite("client_test.db", time.Second*5)
}

func (ts *TextSqlite) clean(ctx context.Context) error {
	newCtx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	_, err := ts.db.ExecContext(newCtx, "DELETE FROM text")
	return err
}

func TestTextSqlite(t *testing.T) {
	suite.Run(t, new(TextSqliteTestSuite))
}

func (ts *TextSqliteTestSuite) SetupTest() {
	ts.Require().NoError(ts.clean(context.Background()))
}

func (ts *TextSqliteTestSuite) TearDownTest() {
	ts.Require().NoError(ts.clean(context.Background()))
}

func (ts *TextSqliteTestSuite) TestSave() {
	err := ts.Save(context.Background(), text1)
	ts.NoError(err)

	saved, err := ts.ByKey(context.Background(), text1.Key)
	ts.NoError(err)
	ts.Equal(text1, saved)
}

func (ts *TextSqliteTestSuite) TestUpdate() {
	textKey1_1 := models.Text{Type: models.TextItem, Tag: "tag1", Key: "key1", Value: "value 1", Comment: "comment", Created: time.Now().Unix()}
	textKey1_2 := models.Text{Type: models.TextItem, Tag: "tag1", Key: "key1", Value: "NEW VALUE", Comment: "NEW COMMENT", Created: time.Now().Unix()}

	err := ts.Save(context.Background(), textKey1_1)
	ts.NoError(err)

	err = ts.Update(context.Background(), textKey1_2)
	ts.NoError(err)

	list, err := ts.All(context.Background())
	ts.NoError(err)
	ts.Equal(1, len(list))
	ts.Equal(textKey1_2, list[0])
}

func (ts *TextSqliteTestSuite) TestByKeyNoRows() {
	text, err := ts.ByKey(context.Background(), "key1")
	ts.ErrorIs(err, ErrItemNotFound)
	ts.Equal(models.Text{}, text)
}

func (ts *TextSqliteTestSuite) TestAll() {
	err := ts.Save(context.Background(), text1)
	ts.NoError(err)
	err = ts.Save(context.Background(), text2)
	ts.NoError(err)

	list, err := ts.All(context.Background())
	ts.NoError(err)
	ts.Equal(2, len(list))

	set := ts.setFromList(list)
	ts.True(set[text1])
	ts.True(set[text2])
}

func (ts *TextSqliteTestSuite) setFromList(list []models.Text) map[models.Text]bool {
	res := make(map[models.Text]bool, len(list))
	for _, text := range list {
		res[text] = true
	}
	return res
}

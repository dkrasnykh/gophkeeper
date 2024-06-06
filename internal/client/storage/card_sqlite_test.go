package storage

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/dkrasnykh/gophkeeper/pkg/models"
)

var (
	card1 = models.Card{Type: models.CardItem, Tag: "tag1", Number: "5106 2110 1025 5079", Exp: "03/25", CVV: 999, Comment: "comment2", Created: time.Now().Unix()}
	card2 = models.Card{Type: models.CardItem, Tag: "tag1", Number: "4149 5678 2364 5978", Exp: "03/26", CVV: 777, Comment: "comment1", Created: time.Now().Unix()}
)

type CardStorager interface {
	All(ctx context.Context) ([]models.Card, error)
	ByNumber(ctx context.Context, number string) (models.Card, error)
	Save(ctx context.Context, card models.Card) error
	Update(ctx context.Context, card models.Card) error
}

type testCardStorager interface {
	CardStorager
	clean(ctx context.Context) error
}

func (s *CardSqlite) clean(ctx context.Context) error {
	newCtx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	_, err := s.db.ExecContext(newCtx, "DELETE FROM card")
	return err
}

type CardSqliteTestSuite struct {
	suite.Suite
	testCardStorager
}

func (ts *CardSqliteTestSuite) SetupSuite() {
	db, _ := New("client_test.db")
	ts.testCardStorager = NewCardSqlite(db, time.Second*5)
}

func TestCardSqlite(t *testing.T) {
	suite.Run(t, new(CardSqliteTestSuite))
}

func (ts *CardSqliteTestSuite) SetupTest() {
	ts.Require().NoError(ts.clean(context.Background()))
}

func (ts *CardSqliteTestSuite) TearDownTest() {
	ts.Require().NoError(ts.clean(context.Background()))
}

func (ts *CardSqliteTestSuite) TestSave() {
	err := ts.Save(context.Background(), card1)
	ts.NoError(err)

	saved, err := ts.ByNumber(context.Background(), card1.Number)
	ts.NoError(err)
	ts.Equal(card1, saved)
}

func (ts *CardSqliteTestSuite) TestUpdate() {
	cardNumber1_1 := models.Card{Type: models.CardItem, Tag: "tag1", Number: "5106 2110 1025 5079", Exp: "03/25", CVV: 999, Comment: "comment2", Created: time.Now().Unix()}
	cardNumber1_2 := models.Card{Type: models.CardItem, Tag: "tag1", Number: "5106 2110 1025 5079", Exp: "03/25", CVV: 999, Comment: "NEW COMMENT", Created: time.Now().Unix()}

	err := ts.Save(context.Background(), cardNumber1_1)
	ts.NoError(err)

	err = ts.Update(context.Background(), cardNumber1_2)
	ts.NoError(err)

	list, err := ts.All(context.Background())
	ts.NoError(err)
	ts.Equal(1, len(list))
	ts.Equal(cardNumber1_2, list[0])
}

func (ts *CardSqliteTestSuite) TestByNumberNoRows() {
	card, err := ts.ByNumber(context.Background(), card1.Number)
	ts.ErrorIs(err, ErrItemNotFound)
	ts.Equal(models.Card{}, card)
}

func (ts *CardSqliteTestSuite) TestAll() {
	err := ts.Save(context.Background(), card1)
	ts.NoError(err)
	err = ts.Save(context.Background(), card2)
	ts.NoError(err)

	list, err := ts.All(context.Background())
	ts.NoError(err)
	ts.Equal(2, len(list))

	set := ts.setFromList(list)
	ts.True(set[card1])
	ts.True(set[card2])
}

func (ts *CardSqliteTestSuite) setFromList(list []models.Card) map[models.Card]bool {
	res := make(map[models.Card]bool, len(list))
	for _, card := range list {
		res[card] = true
	}
	return res
}

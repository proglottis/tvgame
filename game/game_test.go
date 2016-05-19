package game

import (
	"bytes"
	"fmt"
	"testing"
)

const (
	badFile      = `A?`
	questionFile = `A?,Apple
B?,Banana
C?,Carrot
D?,Date
F?,Fig
G?,Grape
K?,Kiwi
L?,Lemon
M?,Mango
O?,Orange`
	questionFileLines = 10
)

type testPlayer struct {
	Name string
}

func (testPlayer) RequestAnswer(question string)                 {}
func (testPlayer) RequestVote(question string, answers []string) {}
func (testPlayer) Complete(game *Game)                           {}

type testHost struct{}

func (testHost) Joined(player Player)                   {}
func (testHost) Question(question *Question)            {}
func (testHost) Vote(question *Question)                {}
func (testHost) Collected(player Player, complete bool) {}
func (testHost) Results(game *Game, results ResultSet)  {}
func (testHost) Complete(game *Game)                    {}

func TestQuestionRepo_bad_format(t *testing.T) {
	buf := bytes.NewBufferString(badFile)
	_, err := NewQuestionRepo(buf)
	if err == nil {
		t.Fatal("Expected bad format error")
	}
}

func TestQuestionRepo_Questions(t *testing.T) {
	repo := newRepo(t)
	for _, test := range []struct {
		N        int
		Expected int
	}{
		{N: 1, Expected: 1},
		{N: questionFileLines, Expected: questionFileLines},
		{N: questionFileLines + 1, Expected: questionFileLines},
	} {
		questions := repo.Questions(make([]*Question, 0, test.N))
		if len(questions) != test.Expected {
			t.Errorf("Expected %d questions, got %d", test.Expected, len(questions))
		}
	}
}

func TestQuestionRepo_Answers(t *testing.T) {
	repo := newRepo(t)
	for _, test := range []struct {
		N        int
		Expected int
	}{
		{N: 1, Expected: 1},
		{N: questionFileLines, Expected: questionFileLines},
		{N: questionFileLines + 1, Expected: questionFileLines},
	} {
		answers := repo.Answers(make([]*Answer, 0, test.N))
		if len(answers) != test.Expected {
			t.Errorf("Expected %d answers, got %d", test.Expected, len(answers))
		}
	}
}

func TestNewResultSet(t *testing.T) {
	p1 := &testPlayer{Name: "B1"}
	p2 := &testPlayer{Name: "B2"}
	a1 := &Answer{Text: "APPLE", Correct: true}
	a2 := &Answer{Text: "BANANA", Player: p1}
	a3 := &Answer{Text: "CARROT", Player: p2}
	question := &Question{
		Text:    "Fruit?",
		Answers: []*Answer{a1, a2, a3},
	}

	results := NewResultSet(question)
	if len(results) > 0 {
		t.Errorf("Expected empty result set, got %#v", results)
	}

	a1.Votes = append(a1.Votes, p1)
	a2.Votes = append(a2.Votes, p2)
	results = NewResultSet(question)
	if results[a1][0].Player != p1 {
		t.Errorf("Expected p1 to get points for correct answer")
	}
	if results[a2][0].Player != p1 {
		t.Errorf("Expected p1 to get points for p2's incorrect answer")
	}
}

func TestAnswerCollector_Collect(t *testing.T) {
	p1 := &testPlayer{}
	p2 := &testPlayer{}
	question := &Question{Text: "Fruit?"}
	collector := AnswerCollector{Question: question, Remaining: 2}

	if err := collector.Collect(p1, " "); err != ErrShortAnswer {
		t.Errorf("Expected ErrShortAnswer, got %s", err)
	}
	if err := collector.Collect(p1, "Apple"); err != nil {
		t.Errorf("Expected success, got %s", err)
	}
	if collector.Remaining != 1 {
		t.Errorf("Expected 1 remaining answer, got %d", collector.Remaining)
	}
	if err := collector.Collect(p1, "Banana"); err != ErrCompleted {
		t.Errorf("Expected ErrCompleted, got %s", err)
	}
	if err := collector.Collect(p2, "Apple"); err != ErrDupAnswer {
		t.Errorf("Expected ErrDupAnswer, got %s", err)
	}
	if err := collector.Collect(p2, "Banana"); err != nil {
		t.Errorf("Expected success, got %s", err)
	}
	if collector.Remaining != 0 {
		t.Errorf("Expected 0 remaining answer, got %d", collector.Remaining)
	}
}

func TestAnswerCollector_Complete(t *testing.T) {
	question := &Question{Text: "Fruit?"}
	collector := AnswerCollector{Question: question, Remaining: 2}
	if collector.Complete() {
		t.Errorf("Expected not to be complete")
	}
	collector.Remaining = 0
	if !collector.Complete() {
		t.Errorf("Expected to be complete")
	}
}

func TestVoteCollector_Collect(t *testing.T) {
	p1 := &testPlayer{}
	p2 := &testPlayer{}
	question := &Question{
		Text: "Fruit?",
		Answers: []*Answer{
			{Text: "APPLE"},
			{Text: "BANANA"},
			{Text: "CARROT", Player: p1},
		},
	}
	collector := VoteCollector{Question: question, Remaining: 2}

	if err := collector.Collect(p1, "Carrot"); err != ErrOwnAnswer {
		t.Errorf("Expected ErrOwnAnswer, got %s", err)
	}
	if err := collector.Collect(p1, "Apple"); err != nil {
		t.Errorf("Expected success, got %s", err)
	}
	if collector.Remaining != 1 {
		t.Errorf("Expected 1 remaining answer, got %d", collector.Remaining)
	}
	if err := collector.Collect(p1, "Banana"); err != ErrCompleted {
		t.Errorf("Expected ErrCompleted, got %s", err)
	}
	if err := collector.Collect(p2, "Nonexistent"); err != ErrNoAnswer {
		t.Errorf("Expected ErrNoAnswer, got %s", err)
	}
	if err := collector.Collect(p2, "Apple"); err != nil {
		t.Errorf("Expected success, got %s", err)
	}
	if collector.Remaining != 0 {
		t.Errorf("Expected 0 remaining answer, got %d", collector.Remaining)
	}
}

func TestVoteCollector_Complete(t *testing.T) {
	question := &Question{Text: "Fruit?"}
	collector := VoteCollector{Question: question, Remaining: 2}
	if collector.Complete() {
		t.Errorf("Expected not to be complete")
	}
	collector.Remaining = 0
	if !collector.Complete() {
		t.Errorf("Expected to be complete")
	}
}

func TestGame_AddPlayer(t *testing.T) {
	host := &testHost{}
	repo := newRepo(t)
	game := New(repo, host)
	for i := 0; i < maxPlayers; i++ {
		if err := game.AddPlayer(&testPlayer{Name: fmt.Sprintf("%d", i+1)}); err != nil {
			t.Fatalf("Expected success, got %s", err)
		}
	}
	if err := game.AddPlayer(&testPlayer{Name: "too many"}); err != ErrRoomFull {
		t.Fatalf("Expected ErrRoomFull, got %s", err)
	}
}

func TestGame_p1_always_wins(t *testing.T) {
	host := &testHost{}
	p1 := &testPlayer{Name: "B1"}
	p2 := &testPlayer{Name: "B2"}
	repo := newRepo(t)
	game := New(repo, host)
	if len(game.Questions) != 7 {
		t.Fatalf("Expected 7 questions")
	}
	game.AddPlayer(p1, p2)
	game.Begin()
	for i := 0; i < 7; i++ {
		if err := game.Collect(p1, "Moose"); err != nil {
			t.Fatalf("%s", err)
		}
		if err := game.Collect(p2, "Monkey"); err != nil {
			t.Fatalf("%s", err)
		}
		game.Vote()
		if err := game.Collect(p1, game.Current().CorrectAnswer().Text); err != nil {
			t.Fatalf("%s", err)
		}
		if err := game.Collect(p2, "Moose"); err != nil {
			t.Fatalf("%s", err)
		}
		game.Stop()
		game.Next()
	}
	if game.Players[p1] != 30000 {
		t.Errorf("p1 expected 30000 points, got %d", game.Players[p1])
	}
	if game.Players[p2] != 0 {
		t.Errorf("p2 expected 0 points, got %d", game.Players[p2])
	}
}

func newRepo(t testing.TB) *QuestionRepo {
	buf := bytes.NewBufferString(questionFile)
	repo, err := NewQuestionRepo(buf)
	if err != nil {
		t.Fatal(err)
	}
	return repo
}

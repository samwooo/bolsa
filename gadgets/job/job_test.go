package job

import (
	"context"
	"fmt"
	"testing"

	"github.com/samwooo/bolsa/gadgets/logging"
	"github.com/samwooo/bolsa/gadgets/util"
	"github.com/stretchr/testify/assert"
)

type batch1WithError struct{}

func (b *batch1WithError) size() int { return 1 }
func (b *batch1WithError) batch(context.Context, []interface{}) (interface{}, error) {
	return nil, fmt.Errorf("batch1 with error")
}

type batch1WithoutError struct{}

func (b *batch1WithoutError) size() int { return 1 }
func (b *batch1WithoutError) batch(ctx context.Context, groupedMash []interface{}) (interface{}, error) {
	return groupedMash[0], nil
}

type batchNWithError struct{ n int }

func (b *batchNWithError) size() int { return b.n }
func (b *batchNWithError) batch(context.Context, []interface{}) (interface{}, error) {
	return nil, fmt.Errorf("batch%d with error", b.n)
}

type batchNWithoutError struct{ n int }

func (b *batchNWithoutError) size() int { return b.n }
func (b *batchNWithoutError) batch(ctx context.Context, groupedMash []interface{}) (interface{}, error) {
	return groupedMash[len(groupedMash)-1], nil
}

type actionWithError struct{}

func (a *actionWithError) act(ctx context.Context, p interface{}) (r interface{}, e error) {
	return nil, fmt.Errorf("action with error")
}

type actionWithoutError struct{}

func (a *actionWithoutError) act(ctx context.Context, p interface{}) (r interface{}, e error) {
	return p, nil
}

// Job Tester
type JobTester struct {
	maxRetry int
	curRetry int
	*Job
}

func (*JobTester) worth(d Done) bool { return d.E != nil }

func (jt *JobTester) forgo() bool {
	ended := jt.curRetry >= jt.maxRetry
	jt.curRetry++
	return ended
}

func newJobTester(bs batchStrategy, as actionStrategy) *JobTester {
	logging.DefaultLogger(fmt.Sprintf(" < %s > ", util.APP_NAME),
		logging.LogLevelFromString("DEBUG"), 100)

	jt := &JobTester{3, 0, NewJob(logging.GetLogger("job test "), 0)}
	jt.BatchWanted(bs).ActionWanted(as).RetryWanted(jt)
	return jt
}

// nil nil
func TestJobWithNoBatchNoAction(t *testing.T) {
	jt := newJobTester(nil, nil)
	with := []interface{}{1, 2, 3}
	allDone := jt.Run(context.Background(), with)
	for _, done := range allDone {
		assert.Equal(t, true, util.IsIn(done.P, with))
		assert.Equal(t, true, util.IsIn(done.R, with))
		assert.Equal(t, nil, done.E)
	}
}

// nil withError
func TestJobWithNoBatchButActionError(t *testing.T) {
	jt := newJobTester(nil, &actionWithError{})
	with := []interface{}{1, 2, 3}
	allDone := jt.Run(context.Background(), with)
	for _, done := range allDone {
		assert.Equal(t, true, util.IsIn(done.P, with))
		assert.Equal(t, nil, done.R)
		assert.Equal(t, "action ×", done.E.Error())
	}
}

// nil withoutError
func TestJobWithNoBatchButAction(t *testing.T) {
	jt := newJobTester(nil, &actionWithoutError{})
	with := []interface{}{1, 2, 3}
	allDone := jt.Run(context.Background(), with)
	for _, done := range allDone {
		assert.Equal(t, true, util.IsIn(done.P, with))
		assert.Equal(t, true, util.IsIn(done.R, with))
		assert.Equal(t, nil, done.E)
	}
}

// 1-withError nil
func TestJobBatch1WithErrorNoAction(t *testing.T) {
	jt := newJobTester(&batch1WithError{}, nil)
	with := []interface{}{1, 2, 3}
	allDone := jt.Run(context.Background(), with)
	for _, done := range allDone {
		assert.Equal(t, true, util.IsIn(done.P, with))
		assert.Equal(t, nil, done.R)
		assert.Equal(t, "batch ×", done.E.Error())
	}
}

// 1-withError withError
func TestJobBatch1WithErrorActionWithError(t *testing.T) {
	jt := newJobTester(&batch1WithError{}, &actionWithError{})
	with := []interface{}{1, 2, 3}
	allDone := jt.Run(context.Background(), with)
	for _, done := range allDone {
		assert.Equal(t, true, util.IsIn(done.P, with))
		assert.Equal(t, nil, done.R)
		assert.Equal(t, "batch ×", done.E.Error())
	}
}

// 1-withError withoutError
func TestJobBatch1WithErrorActionWithoutError(t *testing.T) {
	jt := newJobTester(&batch1WithError{}, &actionWithoutError{})
	with := []interface{}{1, 2, 3}
	allDone := jt.Run(context.Background(), with)
	for _, done := range allDone {
		assert.Equal(t, true, util.IsIn(done.P, with))
		assert.Equal(t, nil, done.R)
		assert.Equal(t, "batch ×", done.E.Error())
	}
}

// 1-withoutError nil
func TestJobBatch1WithoutErrorNoAction(t *testing.T) {
	jt := newJobTester(&batch1WithoutError{}, nil)
	with := []interface{}{1, 2, 3}
	allDone := jt.Run(context.Background(), with)
	for _, done := range allDone {
		assert.Equal(t, true, util.IsIn(done.P, with))
		assert.Equal(t, true, util.IsIn(done.R, with))
		assert.Equal(t, nil, done.E)
	}
}

// 1-withoutError withError
func TestJobBatch1WithoutErrorActionWithError(t *testing.T) {
	jt := newJobTester(&batch1WithoutError{}, &actionWithError{})
	with := []interface{}{1, 2, 3}
	allDone := jt.Run(context.Background(), with)
	for _, done := range allDone {
		assert.Equal(t, true, util.IsIn(done.P, with))
		assert.Equal(t, nil, done.R)
		assert.Equal(t, "action ×", done.E.Error())
	}
}

// 1-withoutError withoutError
func TestJobBatch1WithoutErrorActionWithoutError(t *testing.T) {
	jt := newJobTester(&batch1WithoutError{}, &actionWithoutError{})
	with := []interface{}{1, 2, 3}
	allDone := jt.Run(context.Background(), with)
	for _, done := range allDone {
		assert.Equal(t, true, util.IsIn(done.P, with))
		assert.Equal(t, true, util.IsIn(done.R, with))
		assert.Equal(t, nil, done.E)
	}
}

// 3-withError nil
func TestJobBatchNWithErrorNoAction(t *testing.T) {
	jt := newJobTester(&batchNWithError{3}, nil)
	with := []interface{}{1, 2, 3, 4, 5, 6, 7, 8}
	allDone := jt.Run(context.Background(), with)
	assert.Equal(t, len(with), len(allDone))
	for _, done := range allDone {
		assert.Equal(t, true, util.IsIn(done.P, with))
		assert.Equal(t, nil, done.R)
		assert.Equal(t, "batch ×", done.E.Error())
	}
}

// 3-withError withError
func TestJobBatchNWithErrorActionWithError(t *testing.T) {
	jt := newJobTester(&batchNWithError{3}, &actionWithError{})
	with := []interface{}{1, 2, 3, 4, 5, 6, 7, 8}
	allDone := jt.Run(context.Background(), with)
	assert.Equal(t, len(with), len(allDone))
	for _, done := range allDone {
		assert.Equal(t, true, util.IsIn(done.P, with))
		assert.Equal(t, nil, done.R)
		assert.Equal(t, "batch ×", done.E.Error())
	}
}

// 3-withError withoutError
func TestJobBatchNWithErrorActionWithoutError(t *testing.T) {
	jt := newJobTester(&batchNWithError{3}, &actionWithoutError{})
	with := []interface{}{1, 2, 3, 4, 5, 6, 7, 8}
	allDone := jt.Run(context.Background(), with)
	assert.Equal(t, len(with), len(allDone))
	for _, done := range allDone {
		assert.Equal(t, true, util.IsIn(done.P, with))
		assert.Equal(t, nil, done.R)
		assert.Equal(t, "batch ×", done.E.Error())
	}
}

// 3-withoutError nil
func TestJobBatchNWithoutErrorNoAction(t *testing.T) {
	jt := newJobTester(&batchNWithoutError{3}, nil)
	with := []interface{}{1, 2, 3, 4, 5, 6, 7, 8}
	allDone := jt.Run(context.Background(), with)
	assert.Equal(t, 3, len(allDone))
	for _, done := range allDone {
		assert.Equal(t, true, util.IsIn(done.P, with))
		assert.Equal(t, true, util.IsIn(done.R, with))
		assert.Equal(t, nil, done.E)
	}
}

// 3-withoutError withError
func TestJobBatchNWithoutErrorActionWithError(t *testing.T) {
	jt := newJobTester(&batchNWithoutError{3}, &actionWithError{})
	with := []interface{}{1, 2, 3, 4, 5, 6, 7, 8}
	allDone := jt.Run(context.Background(), with)
	assert.Equal(t, 3, len(allDone))
	for _, done := range allDone {
		assert.Equal(t, true, util.IsIn(done.P, with))
		assert.Equal(t, nil, done.R)
		assert.Equal(t, "action ×", done.E.Error())
	}
}

// 3-withoutError withoutError
func TestJobBatchNWithoutErrorActionWithoutError(t *testing.T) {
	jt := newJobTester(&batchNWithoutError{3}, &actionWithoutError{})
	with := []interface{}{1, 2, 3, 4, 5, 6, 7, 8}
	allDone := jt.Run(context.Background(), with)
	assert.Equal(t, 3, len(allDone))
	for _, done := range allDone {
		assert.Equal(t, true, util.IsIn(done.P, with))
		assert.Equal(t, true, util.IsIn(done.R, with))
		assert.Equal(t, nil, done.E)
	}
}

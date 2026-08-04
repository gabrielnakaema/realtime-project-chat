package apphost

import (
	"errors"
	"io"
	"log/slog"
	"testing"
)

type fakeApplication struct {
	served     bool
	closeCalls int
	serveErr   error
}

func (a *fakeApplication) Serve() error {
	a.served = true
	return a.serveErr
}

func (a *fakeApplication) Close() {
	a.closeCalls++
}

func TestRunReturnsConstructionErrorWithoutServing(t *testing.T) {
	constructionErr := errors.New("construction failed")
	app := &fakeApplication{}

	err := Run("test-service", func() (*fakeApplication, error) {
		return app, constructionErr
	})

	if !errors.Is(err, constructionErr) {
		t.Fatalf("Run() error = %v, want wrapped construction error", err)
	}
	if app.served {
		t.Fatal("Run() served an application whose construction failed")
	}
}

func TestRunServesAndClosesApplication(t *testing.T) {
	app := &fakeApplication{}

	err := Run("test-service", func() (*fakeApplication, error) {
		return app, nil
	})

	if err != nil {
		t.Fatalf("Run() error = %v, want nil", err)
	}
	if !app.served {
		t.Fatal("Run() did not serve the application")
	}
	if app.closeCalls != 1 {
		t.Fatalf("Close() calls = %d, want 1", app.closeCalls)
	}
}

func TestRunClosesApplicationWhenServeFails(t *testing.T) {
	serveErr := errors.New("serve failed")
	app := &fakeApplication{serveErr: serveErr}

	err := Run("test-service", func() (*fakeApplication, error) {
		return app, nil
	})

	if !errors.Is(err, serveErr) {
		t.Fatalf("Run() error = %v, want wrapped serve error", err)
	}
	if !app.served {
		t.Fatal("Run() did not serve the application")
	}
	if app.closeCalls != 1 {
		t.Fatalf("Close() calls = %d, want 1", app.closeCalls)
	}
}

func TestRuntimeCloseCancelsContextAndClosesTrackedResourcesOnce(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	rt := NewForTest(log)
	firstCloseCalls := 0
	secondCloseCalls := 0

	rt.TrackFunc(func() error {
		firstCloseCalls++
		return nil
	})
	rt.TrackFunc(func() error {
		secondCloseCalls++
		return nil
	})

	rt.Close()
	rt.Close()

	if firstCloseCalls != 1 {
		t.Fatalf("first resource Close() calls = %d, want 1", firstCloseCalls)
	}
	if secondCloseCalls != 1 {
		t.Fatalf("second resource Close() calls = %d, want 1", secondCloseCalls)
	}
	select {
	case <-rt.Ctx.Done():
	default:
		t.Fatal("Runtime.Close() did not cancel the runtime context")
	}
}

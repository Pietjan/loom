package styles_test

import (
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"testing"

	"github.com/pietjan/loom/internal/styles"
)

func TestBuilder(t *testing.T) {
	var b styles.Builder
	b.Add("inline-flex items-center")
	b.If(true, "gap-2")
	b.If(false, "hidden")
	styles.Match(&b, "primary", map[string]string{
		"primary": "bg-accent text-accent-content",
		"ghost":   "bg-transparent",
	})

	got := b.String()
	for _, want := range []string{"inline-flex", "items-center", "gap-2", "bg-accent", "text-accent-content"} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in %q", want, got)
		}
	}
	if strings.Contains(got, "hidden") || strings.Contains(got, "bg-transparent") {
		t.Fatalf("unexpected classes in %q", got)
	}
}

func TestMergeUserWins(t *testing.T) {
	got := styles.Merge("px-4 py-2 bg-accent", "bg-red-500")
	if strings.Contains(got, "bg-accent") {
		t.Fatalf("component class should lose the conflict: %q", got)
	}
	if !strings.Contains(got, "bg-red-500") {
		t.Fatalf("user class missing: %q", got)
	}
	if !strings.Contains(got, "px-4") || !strings.Contains(got, "py-2") {
		t.Fatalf("non-conflicting classes must survive: %q", got)
	}
}

func TestMergeEmptyUser(t *testing.T) {
	if got := styles.Merge("px-4", ""); got != "px-4" {
		t.Fatalf("got %q", got)
	}
}

func TestMergeConcurrent(t *testing.T) {
	if os.Getenv("LOOM_MERGE_CONCURRENT_HELPER") == "1" {
		runConcurrentMergeChecks(t)
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=^TestMergeConcurrent$")
	cmd.Env = append(os.Environ(), "LOOM_MERGE_CONCURRENT_HELPER=1")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("concurrent merge subprocess failed: %v\n%s", err, out)
	}
	if len(out) != 0 {
		t.Logf("concurrent merge subprocess output:\n%s", out)
	}
}

func runConcurrentMergeChecks(t *testing.T) {
	t.Helper()

	testCases := []struct {
		component string
		user      string
		mustHave  []string
		mustMiss  []string
	}{
		{
			component: "px-4 py-2 bg-accent text-sm",
			user:      "bg-red-500 p-8",
			mustHave:  []string{"bg-red-500", "p-8", "text-sm"},
			mustMiss:  []string{"bg-accent", "px-4", "py-2"},
		},
		{
			component: "inline-flex items-center gap-2 rounded-lg border border-base-200 px-3 py-2",
			user:      "border-base-400 px-8 hover:bg-base-50",
			mustHave:  []string{"inline-flex", "items-center", "gap-2", "rounded-lg", "border", "border-base-400", "px-8", "py-2", "hover:bg-base-50"},
			mustMiss:  []string{"border-base-200", "px-3"},
		},
		{
			component: "absolute start-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2 text-xs",
			user:      "text-sm dark:text-white",
			mustHave:  []string{"absolute", "start-1/2", "top-1/2", "-translate-x-1/2", "-translate-y-1/2", "text-sm", "dark:text-white"},
			mustMiss:  []string{"text-xs"},
		},
	}

	workers := max(8, runtime.GOMAXPROCS(0)*4)
	iterations := 200
	start := make(chan struct{})
	var wg sync.WaitGroup
	errCh := make(chan string, workers*iterations*len(testCases))

	for range workers {
		wg.Go(func() {
			<-start
			for range iterations {
				for _, tc := range testCases {
					got := styles.Merge(tc.component, tc.user)
					for _, want := range tc.mustHave {
						if !strings.Contains(got, want) {
							errCh <- "missing " + want + " in " + got
						}
					}
					for _, miss := range tc.mustMiss {
						if strings.Contains(got, miss) {
							errCh <- "unexpected " + miss + " in " + got
						}
					}
					if got != styles.Sort(got) {
						errCh <- "non-canonical output " + got
					}
				}
			}
		})
	}

	close(start)
	wg.Wait()
	close(errCh)

	for got := range errCh {
		t.Fatalf("Merge returned invalid output: %s", got)
	}
}

func TestSortStable(t *testing.T) {
	// Same set, different input order -> same output.
	a := styles.Sort("text-sm bg-accent flex px-4 hover:bg-accent/90")
	b := styles.Sort("hover:bg-accent/90 px-4 flex bg-accent text-sm")
	if a != b {
		t.Fatalf("sort not canonical:\n%q\n%q", a, b)
	}
	// Layout before spacing before background.
	flex, spacing, background := strings.Index(a, "flex"), strings.Index(a, "px-4"), strings.Index(a, "bg-accent")
	if flex >= spacing || spacing >= background {
		t.Fatalf("unexpected order: %q", a)
	}
	// Unvarianted before varianted.
	if strings.Index(a, "hover:bg-accent/90") < strings.Index(a, "bg-accent") {
		t.Fatalf("variant should sort after plain utilities: %q", a)
	}
}

func TestSortDeduplicates(t *testing.T) {
	if got := styles.Sort("flex flex px-4"); got != "flex px-4" {
		t.Fatalf("got %q", got)
	}
}

// TestSortIsTotalOrder guards against golden flakiness: tailwind-merge's
// output order is not stable across processes, so Sort must produce the
// same result for any input permutation - including classes that match no
// sort prefix (negative -translate-*) or share an arbitrary variant.
func TestSortIsTotalOrder(t *testing.T) {
	perms := [][]string{
		{"-translate-x-1/2", "-translate-y-1/2", "absolute", "size-5"},
		{"-translate-y-1/2", "size-5", "-translate-x-1/2", "absolute"},
		{"size-5", "absolute", "-translate-y-1/2", "-translate-x-1/2"},
		{"absolute", "-translate-x-1/2", "size-5", "-translate-y-1/2"},
	}
	want := styles.Sort(strings.Join(perms[0], " "))
	for _, p := range perms[1:] {
		if got := styles.Sort(strings.Join(p, " ")); got != want {
			t.Fatalf("Sort not permutation-invariant:\n%q\n%q", want, got)
		}
	}
}

func TestSortKeepsArbitraryValuesIntact(t *testing.T) {
	in := "[&_svg]:size-4 grid-cols-[1fr_auto] supports-[anchor-name:--a]:absolute"
	got := styles.Sort(in)
	for tok := range strings.FieldsSeq(in) {
		if !strings.Contains(got, tok) {
			t.Fatalf("token %q mangled; got %q", tok, got)
		}
	}
}

func BenchmarkMerge(b *testing.B) {
	const component = "inline-flex items-center gap-2 rounded-lg border border-base-200 bg-white px-3 py-2 text-sm text-base-800 dark:border-base-600 dark:bg-base-700 dark:text-base-100"
	const user = "border-base-400 bg-base-50 px-8 text-base hover:bg-base-100 dark:bg-base-800"

	b.ReportAllocs()
	for b.Loop() {
		_ = styles.Merge(component, user)
	}
}

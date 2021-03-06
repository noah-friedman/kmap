package kmap_test

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"github.com/cucumber/godog"
	"github.com/noah-friedman/kmap"
	"math"
	"math/rand"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"testing"
)

func iInitializeTheKmap(ctx context.Context) context.Context {
	k, err := kmap.NewKmap(ctx.Value("size").(int), ctx.Value("args").([]int), ctx.Value("dontCare").([]int))

	return context.WithValue(context.WithValue(ctx, "kmap", k), "err", err)
}

func iRandomlyGenerate(ctx context.Context, key string) context.Context {
	size := int(math.Pow(2, float64(ctx.Value("size").(int))))

	var a []int
	for i := 0; i < size; i++ {
		if rand.Int31n(2) == 1 {
			a = append(a, i)
		}
	}

	return context.WithValue(ctx, key, a)
}
func iRandomlyGenerateTheArgumentsToTheKmap(ctx context.Context) context.Context {
	return iRandomlyGenerate(ctx, "args")
}
func iRandomlyGenerateTheDontCareConditionsOfTheKmap(ctx context.Context) context.Context {
	return iRandomlyGenerate(ctx, "dontCare")
}

func parseTable(t *godog.Table) ([]int, error) {
	var r []int
	for _, v := range t.Rows[0].Cells {
		if i, e := strconv.Atoi(v.Value); e != nil {
			return nil, e
		} else {
			r = append(r, i)
		}
	}

	return r, nil
}
func theArgumentsToTheKmapAre(ctx context.Context, t *godog.Table) (context.Context, error) {
	if args, e := parseTable(t); e != nil {
		return ctx, e
	} else {
		return context.WithValue(ctx, "args", args), nil
	}
}
func theDontCareConditionsOfTheKmapAre(ctx context.Context, t *godog.Table) (context.Context, error) {
	if dontCare, e := parseTable(t); e != nil {
		return ctx, e
	} else {
		return context.WithValue(ctx, "dontCare", dontCare), nil
	}
}

func theKmapSizeIs(ctx context.Context, size int) context.Context {
	return context.WithValue(ctx, "size", size)
}

func swap23(x *int) {
	if *x == 2 {
		*x = 3
	} else if *x == 3 {
		*x = 2
	}
}
func theKmapValuesShouldMatch(ctx context.Context, a string, b bool) error {
	var (
		k        = ctx.Value("kmap").(*kmap.Kmap)
		actual   = make([]*bool, k.Rows*k.Cols)
		expected = ctx.Value(a).([]int)
	)

	for i, v := range k.Values {
		swap23(&i)

		for a, v := range v {
			swap23(&a)

			actual[i*k.Cols+a] = v
		}
	}

	for _, v := range expected {
		if actual := actual[v]; actual == nil || *actual != b {
			return fmt.Errorf("minterm %d: expected &%t, got %v", v, b, actual)
		}
	}

	return nil
}

func theKmapValuesShouldMatchTheArguments(ctx context.Context) error {
	return theKmapValuesShouldMatch(ctx, "args", true)
}
func theKmapValuesShouldMatchTheDontCareConditions(ctx context.Context) error {
	return theKmapValuesShouldMatch(ctx, "dontCare", false)
}
func theKmapValuesShouldMatchTheArgumentsAndDontCareConditions() godog.Steps {
	return godog.Steps{
		"the k-map values should match the arguments",
		"the k-map values should match the don't care conditions",
	}
}

func thePropertyOfTheKmapShouldBe(ctx context.Context, prop string, expected int64) error {
	if actual := reflect.ValueOf(*(ctx.Value("kmap").(*kmap.Kmap))).FieldByName(prop).Int(); expected != actual {
		return fmt.Errorf("expected %d, got %d", expected, actual)
	}

	return nil
}

func anErrorShouldHaveOccurred(ctx context.Context) (context.Context, error) {
	if ctx.Value("err") == nil {
		return ctx, fmt.Errorf("expected an error occurred but found no error")
	}

	return context.WithValue(ctx, "err", error(nil)), nil
}

func iParseTheString(ctx context.Context, s string) context.Context {
	parsed, err := kmap.Parse(s, ctx.Value("delim").(string))

	return context.WithValue(context.WithValue(ctx, "parsed", parsed), "err", err)
}

func theParsingResultShouldBe(ctx context.Context, expected *godog.Table) error {
	parsed := ctx.Value("parsed").([]int)

	for i, v := range expected.Rows[0].Cells {
		if a, e := strconv.Atoi(v.Value); e != nil {
			return e
		} else if a != parsed[i] {
			return fmt.Errorf("index %d: expected %d, got %d", i, parsed[i], a)
		}
	}

	return nil
}

func theDelimiterIs(ctx context.Context, delim string) context.Context {
	return context.WithValue(ctx, "delim", delim)
}

func iCreateTheOutputForTheGeneratedKmap(ctx context.Context) context.Context {
	formatted := ctx.Value("kmap").(*kmap.Kmap).Format()
	return context.WithValue(ctx, "formatted", formatted)
}

func theFormattedOutputShouldMatch(ctx context.Context, expected *godog.DocString) error {
	if actual := ctx.Value("kmap").(*kmap.Kmap).Format(); actual != expected.Content {
		return fmt.Errorf("\nexpected:\n%s\nactual:\n%s\n", expected.Content, actual)
	}

	return nil
}

func iAnswer(ctx context.Context, ans string) error {
	_, e := ctx.Value("input").(*os.File).WriteString(ans + "\n")
	return e
}

func iAnswerTheRandomlyGeneratedArgumentsSeperatedBy(ctx context.Context, delim string) error {
	var ans string

	for _, v := range ctx.Value("args").([]int) {
		ans += fmt.Sprintf("%d%s", v, delim)
	}

	return iAnswer(ctx, strings.TrimSuffix(ans, delim))
}

func iRunTheProgram(ctx context.Context) (context.Context, error) {
	if r, input, e := os.Pipe(); e != nil {
		return ctx, e
	} else if out, w, e := os.Pipe(); e != nil {
		return ctx, e
	} else {
		var (
			output    = make(chan string)
			exitCode  = make(chan int, 1)
			exitError = make(chan error, 1)
		)

		go func() {
			var wg sync.WaitGroup

			wg.Add(2)

			go func() {
				defer wg.Done()

				rdr := bufio.NewReader(out)

				for s, e := rdr.ReadString('\n'); e == nil; s, e = rdr.ReadString('\n') {
					output <- s
				}
			}()

			go func() {
				defer wg.Done()

				c, e := kmap.Program(r, w)

				exitCode <- c
				exitError <- e

				close(exitCode)
				close(exitError)

				_ = w.Close()
			}()

			wg.Wait()

			_ = r.Close()
			_ = input.Close()
			_ = out.Close()

			close(output)
		}()

		return context.WithValue(context.WithValue(context.WithValue(context.WithValue(ctx, "input", input), "output", output), "exitCode", exitCode), "exitError", exitError), nil
	}
}

func iShouldBeAsked(ctx context.Context, expected string) error {
	if actual := strings.TrimSuffix(<-ctx.Value("output").(chan string), "\n"); expected != actual {
		return fmt.Errorf("expected %s, got %s", expected, actual)
	}

	return nil
}

func theProgramShouldOutputAnEmptyKmapOfSize(ctx context.Context, expected int) error {
	var output string
	for v := range ctx.Value("output").(chan string) {
		output += v
	}

	expected = int(math.Pow(2, float64(expected)))

	if actual := strings.Count(output, "0"); expected != actual {
		return fmt.Errorf("expected %d cells, found %d", expected, actual)
	}

	return nil
}

func theParsingResultShouldBeEmpty(ctx context.Context) error {
	if len(ctx.Value("parsed").([]int)) != 0 {
		return fmt.Errorf("expected empty parsing result, but length > 0")
	}

	return nil
}

func theProgramShouldExitCleanly(ctx context.Context) error {
	if actual := <-ctx.Value("exitCode").(chan int); actual != 0 {
		return fmt.Errorf("exitted with code %d, error: '%s'", actual, (<-ctx.Value("exitError").(chan error)).Error())
	}

	return nil
}

func iSetTheCommandlineArgumentTo(arg, val string) {
	os.Args = append(os.Args, "-"+arg, val)
}

func theProgramShouldOutput(ctx context.Context, expected *godog.DocString) error {
	output := ctx.Value("output").(chan string)

	var actual string
	for v := range output {
		actual += v
	}
	actual = strings.TrimSuffix(actual, "\n")

	if expected.Content != actual {
		return fmt.Errorf("\nexpected:\n%s\nactual:\n%s\n", expected.Content, actual)
	}

	return nil
}

func iRandomlyGenerateTheArgumentsAndDontCareConditionsForTheKmap(ctx context.Context) context.Context {
	size := int(math.Pow(2, float64(ctx.Value("size").(int))))

	var args, dontCare []int
	for i := 0; i < size; i++ {
		switch rand.Intn(3) {
		case 1:
			args = append(args, i)
		case 2:
			dontCare = append(dontCare, i)
		}
	}

	return context.WithValue(context.WithValue(ctx, "args", args), "dontCare", dontCare)
}

var initialState = map[string]interface{}{
	"kmap":      (*kmap.Kmap)(nil),
	"size":      3,
	"args":      []int(nil),
	"err":       error(nil),
	"parsed":    []int(nil),
	"delim":     "",
	"formatted": "",
	"input":     (*os.File)(nil),
	"output":    (chan string)(nil),
	"exitCode":  (chan int)(nil),
	"exitError": (chan error)(nil),
	"dontCare":  []int(nil),
}

func Stepdefs(ctx *godog.ScenarioContext) {
	ctx.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
		for k, v := range initialState {
			ctx = context.WithValue(ctx, k, v)
		}

		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
		os.Args = os.Args[0:1]

		return ctx, nil
	})

	ctx.Step(`^I initialize the k-map$`, iInitializeTheKmap)
	ctx.Step(`^I randomly generate the arguments to the k-map$`, iRandomlyGenerateTheArgumentsToTheKmap)
	ctx.Step(`^I randomly generate the don\'t care conditions of the k-map$`, iRandomlyGenerateTheDontCareConditionsOfTheKmap)
	ctx.Step(`^the arguments to the k-map are$`, theArgumentsToTheKmapAre)
	ctx.Step(`^the don\'t care conditions of the k-map are$`, theDontCareConditionsOfTheKmapAre)
	ctx.Step(`^the k-map size is (\d+)$`, theKmapSizeIs)
	ctx.Step(`^the k-map values should match the arguments$`, theKmapValuesShouldMatchTheArguments)
	ctx.Step(`^the k-map values should match the don\'t care conditions$`, theKmapValuesShouldMatchTheDontCareConditions)
	ctx.Step(`^the k-map values should match the arguments and don\'t care conditions$`, theKmapValuesShouldMatchTheArgumentsAndDontCareConditions)
	ctx.Step(`^the "([^"]*)" property of the k-map should be (\d+)$`, thePropertyOfTheKmapShouldBe)
	ctx.Step(`^an error should have occurred$`, anErrorShouldHaveOccurred)
	ctx.Step(`^I parse the string "([^"]*)"$`, iParseTheString)
	ctx.Step(`^the parsing result should be$`, theParsingResultShouldBe)
	ctx.Step(`^the delimiter is "([^"]*)"$`, theDelimiterIs)
	ctx.Step(`^I create the output for the generated k-map$`, iCreateTheOutputForTheGeneratedKmap)
	ctx.Step(`^the formatted output should match$`, theFormattedOutputShouldMatch)
	ctx.Step(`^I answer "([^"]*)"$`, iAnswer)
	ctx.Step(`^I answer the randomly generated arguments seperated by "([^"]*)"$`, iAnswerTheRandomlyGeneratedArgumentsSeperatedBy)
	ctx.Step(`^I run the program$`, iRunTheProgram)
	ctx.Step(`^I should be asked "([^"]*)"$`, iShouldBeAsked)
	ctx.Step(`^the program should output an empty k-map of size (\d+)$`, theProgramShouldOutputAnEmptyKmapOfSize)
	ctx.Step(`^the parsing result should be empty$`, theParsingResultShouldBeEmpty)
	ctx.Step(`^the program should exit cleanly$`, theProgramShouldExitCleanly)
	ctx.Step(`^I set the "([^"]*)" command-line argument to "([^"]*)"$`, iSetTheCommandlineArgumentTo)
	ctx.Step(`^the program should output$`, theProgramShouldOutput)
	ctx.Step(`^I randomly generate the arguments and don\'t care conditions for the k-map$`, iRandomlyGenerateTheArgumentsAndDontCareConditionsForTheKmap)

	ctx.StepContext().After(func(ctx context.Context, _ *godog.Step, status godog.StepResultStatus, err error) (context.Context, error) {
		if status == godog.StepFailed {
			r := "CONTEXT:\n"

			for k := range initialState {
				r += fmt.Sprintf("%s: %v\n", k, ctx.Value(k))
			}

			return ctx, fmt.Errorf(r)
		}

		return ctx, nil
	})
}

func TestFeatures(t *testing.T) {
	if r := (godog.TestSuite{
		ScenarioInitializer: Stepdefs,
		Options: &godog.Options{
			Concurrency: 1,
			Format:      "pretty",
			Paths:       []string{"features"},
			Randomize:   -1,
			TestingT:    t,
			Tags:        "~@wip",
		},
	}).Run(); r != 0 {
		t.Fatalf("godog exited with non-zero exit code '%d'", r)
	}
}

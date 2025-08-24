package main

import (
	"fmt"
	"os/exec"
)

func main() {
	fmt.Printf("🗑️  ZeroUI Batch Dependency Removal\n")
	fmt.Printf("==================================\n\n")

	// Low-risk dependencies that are safe to remove
	lowRiskDeps := []string{
		// Batch 1 - Static analysis and code quality tools
		"4d63.com/gocheckcompilerdirectives",
		"4d63.com/gochecknoglobals",
		"github.com/Antonboom/errname",
		"github.com/Antonboom/nilnil",
		"github.com/Antonboom/testifylint",
		"github.com/Crocmagnon/fatcontext",
		"github.com/Djarvur/go-err113",
		"github.com/GaijinEntertainment/go-exhaustruct/v3",
		"github.com/KyleBanks/depth",
		"github.com/Masterminds/semver/v3",
		"github.com/OpenPeeDeeP/depguard/v2",
		"github.com/PuerkitoBio/purell",
		"github.com/PuerkitoBio/urlesc",
		"github.com/aclements/go-moremath",
		"github.com/alecthomas/chroma/v2",
		"github.com/alecthomas/go-check-sumtype",
		"github.com/alexkohler/nakedret/v2",
		"github.com/alexkohler/prealloc",
		"github.com/alingse/asasalint",
		"github.com/alingse/nilnesserr",
		"github.com/ashanbrown/forbidigo",
		"github.com/ashanbrown/makezero",
		"github.com/atotto/clipboard",
		"github.com/aymanbagabas/go-osc52/v2",
		"github.com/aymerick/douceur",
		"github.com/beorn7/perks",
		"github.com/bep/godartsass/v2",
		"github.com/bep/golibsass",
		"github.com/bitfield/gotestdox",
		"github.com/bkielbasa/cyclop",
		"github.com/blizzy78/varnamelen",
		"github.com/bombsimon/wsl/v4",
		"github.com/breml/bidichk",
		"github.com/breml/errchkjson",
		"github.com/butuzov/ireturn",
		"github.com/butuzov/mirror",
		"github.com/catenacyber/perfsprint",
		"github.com/catppuccin/go",
		"github.com/ccojocar/zxcvbn-go",
		"github.com/cespare/xxhash/v2",
		"github.com/charithe/durationcheck",
		"github.com/chavacava/garif",
		"github.com/chzyer/readline",
		"github.com/ckaznocha/intrange",
		"github.com/cpuguy83/go-md2man/v2",
		"github.com/creack/pty",
		"github.com/curioswitch/go-reassign",
		"github.com/daixiang0/gci",
		"github.com/davecgh/go-spew",
		"github.com/denis-tingaikin/go-header",
		"github.com/dlclark/regexp2",
		"github.com/dnephin/pflag",
		"github.com/dustin/go-humanize",
		"github.com/erikgeiser/coninput",
		"github.com/ettle/strcase",
		"github.com/fatih/color",
		"github.com/fatih/structtag",
		"github.com/firefart/nonamedreturns",
		"github.com/fzipp/gocyclo",
		"github.com/gabriel-vasile/mimetype",
		"github.com/ghostiam/protogetter",
		"github.com/go-critic/go-critic",
		"github.com/go-logfmt/logfmt",
		"github.com/go-openapi/jsonpointer",
		"github.com/go-openapi/jsonreference",
		"github.com/go-openapi/spec",
		"github.com/go-openapi/swag",
		"github.com/go-playground/locales",
		"github.com/go-playground/universal-translator",
		"github.com/go-toolsmith/astcast",
		"github.com/go-toolsmith/astcopy",
		"github.com/go-toolsmith/astequal",
		"github.com/go-toolsmith/astfmt",
		"github.com/go-toolsmith/astp",
		"github.com/go-toolsmith/strparse",
		"github.com/go-toolsmith/typep",
		"github.com/go-viper/mapstructure/v2",
		"github.com/go-xmlfmt/xmlfmt",
		"github.com/gobwas/glob",
		"github.com/gofrs/flock",
		"github.com/gohugoio/hugo",
		"github.com/golang/mock",
		"github.com/golang/protobuf",
		"github.com/google/go-cmp",
		"github.com/google/shlex",
		"github.com/gordonklaus/ineffassign",
		"github.com/gorilla/css",
		"github.com/gostaticanalysis/analysisutil",
		"github.com/gostaticanalysis/comment",
		"github.com/gostaticanalysis/forcetypeassert",
		"github.com/gostaticanalysis/nilerr",
		"github.com/hashicorp/go-immutable-radix/v2",
		"github.com/hashicorp/go-version",
		"github.com/hashicorp/yamux",
		"github.com/hexops/gotextdiff",
		"github.com/ianlancetaylor/demangle",
		"github.com/inconshreveable/mousetrap",
		"github.com/jgautheron/goconst",
		"github.com/jingyugao/rowserrcheck",
		"github.com/jjti/go-spancheck",
		"github.com/josharian/intern",
		"github.com/julz/importas",
		"github.com/karamaru-alpha/copyloopvar",
		"github.com/kisielk/errcheck",
		"github.com/kkHAIKE/contextcheck",
		"github.com/knadh/koanf/maps",
		"github.com/kulti/thelper",
		"github.com/kunwardeep/paralleltest",
		"github.com/lasiar/canonicalheader",
		"github.com/ldez/exptostd",
		"github.com/ldez/gomoddirectives",
		"github.com/ldez/grignotin",
		"github.com/ldez/tagliatelle",
		"github.com/ldez/usetesting",
		"github.com/leodido/go-urn",
		"github.com/leonklingele/grouper",
		"github.com/lucasb-eyer/go-colorful",
		"github.com/macabu/inamedparam",
		"github.com/mailru/easyjson",
		"github.com/maratori/testableexamples",
		"github.com/maratori/testpackage",
		"github.com/matoous/godox",
		"github.com/mattn/go-colorable",
		"github.com/mattn/go-isatty",
		"github.com/mattn/go-localereader",
		"github.com/mattn/go-runewidth",
		"github.com/mgechev/revive",
		"github.com/microcosm-cc/bluemonday",
		"github.com/mitchellh/copystructure",
		"github.com/mitchellh/go-homedir",
		"github.com/mitchellh/hashstructure/v2",
		"github.com/mitchellh/mapstructure",
		"github.com/mitchellh/reflectwalk",
		"github.com/moricho/tparallel",
		"github.com/muesli/ansi",
		"github.com/muesli/cancelreader",
		"github.com/muesli/reflow",
		"github.com/muesli/termenv",
		"github.com/munnerz/goautoneg",
		"github.com/nakabonne/nestif",
		"github.com/nishanths/exhaustive",
		"github.com/nishanths/predeclared",
		"github.com/nunnatsa/ginkgolinter",
		"github.com/oklog/run",
		"github.com/olekukonko/tablewriter",
		"github.com/pelletier/go-toml",
		"github.com/pelletier/go-toml/v2",
		"github.com/pmezard/go-difflib",
		"github.com/polyfloyd/go-errorlint",
		"github.com/quasilyte/go-ruleguard",
		"github.com/quasilyte/go-ruleguard/dsl",
		"github.com/quasilyte/gogrep",
		"github.com/quasilyte/regex/syntax",
		"github.com/quasilyte/stdinfo",
		"github.com/raeperd/recvcheck",
		"github.com/rivo/uniseg",
		"github.com/rogpeppe/go-internal",
		"github.com/russross/blackfriday/v2",
		"github.com/ryancurrah/gomodguard",
		"github.com/ryanrolds/sqlclosecheck",
		"github.com/sagikazarmark/locafero",
		"github.com/sahilm/fuzzy",
		"github.com/sanposhiho/wastedassign/v2",
		"github.com/santhosh-tekuri/jsonschema/v6",
		"github.com/sashamelentyev/interfacebloat",
		"github.com/sashamelentyev/usestdlibvars",
		"github.com/securego/gosec/v2",
		"github.com/sivchari/containedctx",
		"github.com/sivchari/tenv",
		"github.com/sonatard/noctx",
		"github.com/sourcegraph/conc",
		"github.com/sourcegraph/go-diff",
		"github.com/spf13/afero",
		"github.com/spf13/cast",
		"github.com/spf13/pflag",
		"github.com/ssgreg/nlreturn/v2",
		"github.com/stbenjam/no-sprintf-host-port",
		"github.com/stretchr/objx",
		"github.com/subosito/gotenv",
		"github.com/swaggo/swag",
		"github.com/tdakkota/asciicheck",
		"github.com/tdewolff/parse/v2",
		"github.com/tetafro/godot",
		"github.com/timakin/bodyclose",
		"github.com/timonwong/loggercheck",
		"github.com/tomarrell/wrapcheck/v2",
		"github.com/tommy-muehle/go-mnd/v2",
		"github.com/ultraware/funlen",
		"github.com/ultraware/whitespace",
		"github.com/urfave/cli/v2",
		"github.com/uudashr/gocognit",
		"github.com/uudashr/iface",
		"github.com/xen0n/gosmopolitan",
		"github.com/xo/terminfo",
		"github.com/yagipy/maintidx",
		"github.com/yeya24/promlinter",
		"github.com/ykadowak/zerologlint",
		"github.com/yuin/goldmark",
		"github.com/yuin/goldmark-emoji",
		"gitlab.com/bosi/decorder",
		"go-simpler.org/musttag",
		"go-simpler.org/sloglint",
		"go.opentelemetry.io/otel",
		"go.opentelemetry.io/otel/sdk/metric",
		"go.yaml.in/yaml/v3",
		"gopkg.in/yaml.v2",
		"mvdan.cc/unparam",
		"sigs.k8s.io/yaml",
	}

	fmt.Printf("📋 Total low-risk dependencies to remove: %d\n\n", len(lowRiskDeps))

	// Process in batches of 10 to avoid overwhelming the system
	batchSize := 10
	removed := 0
	failed := 0

	for i := 0; i < len(lowRiskDeps); i += batchSize {
		end := i + batchSize
		if end > len(lowRiskDeps) {
			end = len(lowRiskDeps)
		}

		batch := lowRiskDeps[i:end]
		fmt.Printf("🗑️  Removing batch %d/%d (%d deps)...\n", i/batchSize+1, (len(lowRiskDeps)+batchSize-1)/batchSize, len(batch))

		for _, dep := range batch {
			fmt.Printf("   Removing: %s\n", dep)

			// Use go get to remove the dependency
			cmd := exec.Command("go", "get", dep+"@none")
			output, err := cmd.CombinedOutput()

			if err != nil {
				fmt.Printf("   ❌ Failed to remove %s: %v\n", dep, err)
				if len(output) > 0 {
					fmt.Printf("      Output: %s\n", string(output))
				}
				failed++
			} else {
				fmt.Printf("   ✅ Removed: %s\n", dep)
				removed++
			}
		}

		// Run go mod tidy after each batch
		fmt.Printf("   🔧 Running go mod tidy...\n")
		tidyCmd := exec.Command("go", "mod", "tidy")
		if tidyOutput, tidyErr := tidyCmd.CombinedOutput(); tidyErr != nil {
			fmt.Printf("   ⚠️  go mod tidy warning: %v\n", tidyErr)
			if len(tidyOutput) > 0 {
				fmt.Printf("      Output: %s\n", string(tidyOutput))
			}
		}
	}

	fmt.Printf("\n📊 Batch Removal Summary:\n")
	fmt.Printf("   ✅ Successfully removed: %d dependencies\n", removed)
	fmt.Printf("   ❌ Failed to remove: %d dependencies\n", failed)
	fmt.Printf("   📈 Success rate: %.1f%%\n", float64(removed)/float64(len(lowRiskDeps))*100)

	fmt.Printf("\n🔍 Final cleanup...\n")
	finalTidy := exec.Command("go", "mod", "tidy")
	if _, finalErr := finalTidy.CombinedOutput(); finalErr != nil {
		fmt.Printf("⚠️  Final go mod tidy warning: %v\n", finalErr)
	} else {
		fmt.Printf("✅ Final go mod tidy completed\n")
	}

	// Check final dependency count
	finalCount := 0
	if countCmd := exec.Command("sh", "-c", "go list -m all | grep -v github.com/mrtkrcm/ZeroUI | wc -l"); countCmd != nil {
		if countOutput, countErr := countCmd.Output(); countErr == nil {
			if _, parseErr := fmt.Sscanf(string(countOutput), "%d", &finalCount); parseErr == nil {
				fmt.Printf("📊 Final dependency count: %d\n", finalCount)
				fmt.Printf("📉 Dependencies removed: %d\n", 519-finalCount)
			}
		}
	}

	fmt.Printf("\n✅ Batch dependency removal complete!\n")
	fmt.Printf("📝 Next steps:\n")
	fmt.Printf("   1. Test your application to ensure everything works\n")
	fmt.Printf("   2. Run tests: go test ./...\n")
	fmt.Printf("   3. Build the application: go build\n")
	fmt.Printf("   4. If issues occur, restore from backup branch\n")
}

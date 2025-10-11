package ingest

import (
	"math/rand"
	"time"

	"github.com/pickjonathan/sdek-cli/pkg/types"
)

// Generator is the interface that all source-specific generators must implement
type Generator interface {
	// Generate creates events for the source with deterministic seeding
	Generate(seed int64, count int) ([]types.Event, error)

	// GetSourceType returns the source type this generator handles
	GetSourceType() string
}

// BaseGenerator provides common functionality for all generators
type BaseGenerator struct {
	rand      *rand.Rand
	startTime time.Time
	endTime   time.Time
}

// NewBaseGenerator creates a new base generator with the given seed
func NewBaseGenerator(seed int64) *BaseGenerator {
	// Create a new random source with the seed for deterministic generation
	source := rand.NewSource(seed)
	rng := rand.New(source)

	// Set time range: events within the last 90 days
	endTime := time.Now()
	startTime := endTime.AddDate(0, 0, -90)

	return &BaseGenerator{
		rand:      rng,
		startTime: startTime,
		endTime:   endTime,
	}
}

// RandomTimestamp generates a random timestamp within the valid range
func (bg *BaseGenerator) RandomTimestamp() time.Time {
	// Calculate the duration between start and end
	duration := bg.endTime.Sub(bg.startTime)

	// Generate a random duration within that range
	randomDuration := time.Duration(bg.rand.Int63n(int64(duration)))

	// Add it to the start time
	return bg.startTime.Add(randomDuration)
}

// RandomInt generates a random integer between min and max (inclusive)
func (bg *BaseGenerator) RandomInt(min, max int) int {
	if min >= max {
		return min
	}
	return min + bg.rand.Intn(max-min+1)
}

// RandomString generates a random string from a list of options
func (bg *BaseGenerator) RandomString(options []string) string {
	if len(options) == 0 {
		return ""
	}
	return options[bg.rand.Intn(len(options))]
}

// RandomBool generates a random boolean with the given probability of true
func (bg *BaseGenerator) RandomBool(probability float64) bool {
	return bg.rand.Float64() < probability
}

// RandomElement returns a random element from a slice
func (bg *BaseGenerator) RandomElement(slice []string) string {
	if len(slice) == 0 {
		return ""
	}
	return slice[bg.rand.Intn(len(slice))]
}

// RandomSubset returns a random subset of the input slice
func (bg *BaseGenerator) RandomSubset(slice []string, minSize, maxSize int) []string {
	if len(slice) == 0 {
		return []string{}
	}

	// Determine subset size
	size := bg.RandomInt(minSize, maxSize)
	if size > len(slice) {
		size = len(slice)
	}

	// Shuffle and take first 'size' elements
	shuffled := make([]string, len(slice))
	copy(shuffled, slice)
	bg.rand.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	return shuffled[:size]
}

// ValidateEventCount ensures the count is within valid boundaries
func ValidateEventCount(count int) error {
	if count < 10 || count > 50 {
		return &EventCountError{Count: count}
	}
	return nil
}

// EventCountError represents an error when event count is out of bounds
type EventCountError struct {
	Count int
}

func (e *EventCountError) Error() string {
	return "event count must be between 10 and 50, got " + string(rune(e.Count))
}

// Common author names for events
var AuthorNames = []string{
	"Alice Johnson",
	"Bob Smith",
	"Carol Williams",
	"David Brown",
	"Eve Davis",
	"Frank Miller",
	"Grace Wilson",
	"Henry Moore",
	"Ivy Taylor",
	"Jack Anderson",
}

// Common branch names for Git
var BranchNames = []string{
	"main",
	"develop",
	"feature/auth",
	"feature/api",
	"bugfix/security",
	"hotfix/critical",
	"release/v1.0",
}

// Common file paths for changes
var FilePaths = []string{
	"src/main.go",
	"src/auth/handler.go",
	"src/api/routes.go",
	"tests/integration_test.go",
	"docs/README.md",
	"config/app.yaml",
	"pkg/security/crypto.go",
	"internal/database/queries.go",
}

// Common commit message prefixes
var CommitPrefixes = []string{
	"feat:",
	"fix:",
	"docs:",
	"style:",
	"refactor:",
	"test:",
	"chore:",
}

// Common security keywords
var SecurityKeywords = []string{
	"authentication",
	"authorization",
	"encryption",
	"audit",
	"compliance",
	"access control",
	"vulnerability",
	"security patch",
	"SSL/TLS",
	"RBAC",
}

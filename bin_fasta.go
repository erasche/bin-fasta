package main

import (
	"flag"
	"fmt"
	genetic "github.com/handcraftsman/GeneticGo"
	"github.com/op/go-logging"
	"io/ioutil"
	"math"
	"strconv"
	"strings"
	"time"
)

var log = logging.MustGetLogger("main")

type resource struct {
	name   string
	length int
}

func main() {
	var lengthTable = flag.String("lengthTable", "", "Source length table (2 columns, name<TAB>length)")
	var targetLength = flag.Int("targetLength", 1000, "Target length for bins")
	var maxBins = flag.Int("maxBins", 10, "Try and have fewer bins than this")
	var batchSize = flag.Int("batchSize", 40, "Batch N items at a time. MUST be <90")
	var slop = flag.Int("slop", 100, "Allow a certain amount of slop.")
	var patience = flag.Int("patience", 0, "Integer 0-5, with the max being Dalai-Lama-level patience")

	flag.Parse()

	resources := []resource{}

	content, err := ioutil.ReadFile(*lengthTable)
	if err != nil {
		//Do something
		panic(err)
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		data := strings.Split(line, "\t")
		if len(data) == 2 {
			length, _ := strconv.Atoi(data[1])
			resources = append(
				resources,
				*&resource{
					name:   data[0],
					length: length,
				},
			)
		}
	}

	geneSet := "qwertyuiopasdfghjklzxcvbnmQWERTYUIOPASDFGHJKLZXCVBNM1234567890-=_+!@#$%^&*()<>?|{}[];:',./\\"[0:*batchSize]

	fmt.Printf("# Round IDX\tBin Idx\tSum\tFeature IDs\n")
	for i := 0; i <= len(resources) / *batchSize; i++ {

		min_bound := i * (*batchSize)
		max_bound := (i + 1) * (*batchSize)
		max_bound = int(math.Min(float64(max_bound), float64(len(resources))))

		localResources := resources[min_bound:max_bound]
		log.Info(fmt.Sprintf("Processing %d items", len(localResources)))
		calc := func(candidate string) int {
			decoded := decodeGenes(candidate, localResources, geneSet)
			return getFitness(localResources, decoded, *targetLength, *maxBins, *slop)
		}
		start := time.Now()
		disp := func(candidate string) {
			decoded := decodeGenes(candidate, localResources, geneSet)
			fitness := getFitness(localResources, decoded, *targetLength, *maxBins, *slop)
			display(localResources, decoded, fitness, time.Since(start), i, false)
		}

		var solver = new(genetic.Solver)
		solver.MaxSecondsToRunWithoutImprovement = 1 + float64(*patience)*20
		solver.MaxRoundsWithoutImprovement = 10 + (*patience)*50

		var best = solver.GetBest(calc, disp, geneSet, *maxBins, 32)
		log.Info("Final:")

		decoded := decodeGenes(best, localResources, geneSet)
		fitness := getFitness(localResources, decoded, *targetLength, *maxBins, *slop)
		display(localResources, decoded, fitness, time.Since(start), i, true)

	}

}

func display(allResources []resource, resourceCounts map[resource]int, fitness int, elapsed time.Duration, roundIdx int, dumpData bool) {

	bins := make(map[int][]resource)
	binsum := make(map[int]int)
	usemap := make(map[string]bool)

	for _, resource := range allResources {
		usemap[resource.name] = false
	}

	for resource, count := range resourceCounts {
		bins[count] = append(bins[count], resource)
		binsum[count] += resource.length
		usemap[resource.name] = true
	}

	avg := 0.0
	for bin_idx, _ := range bins {
		avg += float64(binsum[bin_idx]) / float64(len(bins))
	}
	unused_count := 0
	for _, used := range usemap {
		if !used {
			unused_count += 1
		}
	}
	if !dumpData {
		log.Debug(fmt.Sprintf(
			"%d\t%d bins averaging %0.3f. Unused %d\t%s",
			fitness,
			len(bins),
			avg,
			unused_count,
			elapsed))
	} else {
		nice_bin_idx := 1
		for bin_idx, bins := range bins {
			resnames := make([]string, 0)
			for _, res := range bins {
				resnames = append(resnames, res.name)
			}
			fmt.Printf("%d\t%d\t%d\t%s\n", roundIdx, nice_bin_idx, binsum[bin_idx], strings.Join(resnames, ","))
			nice_bin_idx++
		}
	}
}

func decodeGenes(candidate string, lresources []resource, geneSet string) map[resource]int {
	resourceCounts := make(map[resource]int, len(candidate)/2)
	for i := 0; i < len(candidate); i += 2 {
		chromosome := candidate[i : i+2]
		resourceId := scale(strings.Index(geneSet, chromosome[0:1]), len(geneSet), len(lresources))
		resourceCount := strings.Index(geneSet, chromosome[1:2])
		resource := lresources[resourceId]
		resourceCounts[resource] = resourceCounts[resource] + resourceCount
	}
	return resourceCounts
}

func scale(value, currentMax, newMax int) int {
	return value * newMax / currentMax
}

func getFitness(allResources []resource, lresources map[resource]int, targetLength int, maxBins int, slop int) int {
	score := 0

	bins := make(map[int]int)
	use_map := make(map[string]bool)

	// Initialize with zeros
	for _, res := range allResources {
		use_map[res.name] = false
	}

	// loop over our resources
	for resource, count := range lresources {
		// separate into bins
		bins[count] += resource.length
		// mark as used when used
		use_map[resource.name] = true

		if count > maxBins {
			score -= count * 100
		}
	}

	// Loop over bins
	for _, value := range bins {
		diff := math.Abs(float64(targetLength - value))
		if diff > float64(slop) {
			score -= int(diff)
		} else {
			score += 1000
		}
	}

	for _, used := range use_map {
		if !used {
			score -= 10000
		}
	}

	return score
}

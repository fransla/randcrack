package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"sync"
)

var (
	buf    bytes.Buffer
	logger = log.New(&buf, "logger: ", log.Lshortfile)
)

// concurrency stuff
var wg sync.WaitGroup
var once sync.Once

type lcg struct {
	multiplier int64
	addend     int64
	mask       int64
	seed       int64
}

func NewLCG() lcg {
	newLCG := lcg{}
	newLCG.multiplier = 0x5DEECE66D
	newLCG.addend = 0xB
	newLCG.mask = ((1 << 48) - 1)
	return newLCG
}

func (l *lcg) updateSeed() {
	l.seed = ((l.seed*l.multiplier + l.addend) & l.mask)
}

func (l *lcg) outputShift(nbits uint) int32 {
	return (int32)(l.seed >> nbits)
}

func (l lcg) mod(a, b int) int {
	return a % b
}

func (l lcg) fixDist(nbits int, val, n int) int {
	return nbits - val + n
}

func (l *lcg) SetSeed(newseed int64) {
	l.seed = (newseed ^ l.multiplier) & l.mask
	return
}

func (l *lcg) Next(nbits uint) int32 {
	l.updateSeed()
	return l.outputShift((48 - nbits))
}

func (l *lcg) NextInt32() int32 {
	return l.Next(32)
}

func (l *lcg) NextInt(n int32) int32 {
	if (n & -n) == n { // i.e., n is a power of 2
		return (int32)(((int64)(n) * int64(l.Next(31))) >> 31)
	}
	var val int32
	for true {
		//update seed
		//1 l.seed = ((l.seed*l.multiplier + l.addend) & l.mask)
		//1 var nbits int32 = (int32)(l.seed >> 17)
		var nbits int32 = l.Next(31)
		val = nbits % n
		if (nbits - val + n) >= 1 {
			break
		}
	}
	return val
}

func Prime_factors2(n int32) int32 {
	//Returns the number of times 2 is a prime factors of a positive integer
	var count2 int32 = 0
	for n%2 == 0 {
		count2++
		n = n / 2
	}
	return count2
}

func crackNextInt(values []string, missing int) []int64 {
	//Tries to crack nextInt() given at least 2 values from the PRNG, with possibly missing values
	if len(values) < 2 {
		return nil
	}

	var toTest = make([]int, len(values))
	var err error
	for i := range values {
		toTest[i], err = strconv.Atoi(strings.TrimSpace(values[i]))
		if err != nil {
			panic(err)
		}
	}

	var result = NewLCG()
	var seedlist []int64
	for i := 0; i < int(math.Pow(2, 16)); i++ {
		result.seed = (int64)(toTest[0])*int64(math.Pow(2, 16)) + (int64)(i)
		//logger.Print("seed: ", result.seed)
		var nomatch bool = false
		var index, lastFound int = 1, 0
		for j := 1; j <= (missing+1)*(len(toTest)-1); j++ {
			var candidate int32 = result.NextInt32()
			//logger.Printf("candidate: %d, toTest[index]: %d ", candidate, (int32)(toTest[index]))
			if candidate == (int32)(toTest[index]) {
				//logger.Printf("index %d and candidate %d ", index, candidate)
				lastFound = j
				index++
			} else if j > lastFound+missing {
				nomatch = true
				break
			}
			if index == len(toTest) {
				break
			}
		}
		if !nomatch {
			logger.Print("found!")
			result.seed = (int64)(toTest[0])*int64(math.Pow(2, 16)) + (int64)(i)
			seedlist = append(seedlist, result.seed)
			//			if not args.full:
			//				break
		}
	}
	if len(seedlist) > 0 {
		return seedlist
	}
	logger.Print("Not found :-(")
	return nil
}

func crackNextIntn(values []string, missing int, n int32) lcg {
	//Tries to crack nextInt(n) given at least 2 values from the PRNG, with possibly missing values, n even but not a power of 2
	var result = NewLCG()
	if len(values) < 2 {
		return result
	}
	var toTest = make([]int, len(values))
	var err error
	for i := range values {
		toTest[i], err = strconv.Atoi(strings.TrimSpace(values[i]))
		if err != nil {
			panic(err)
		}
	}
	var seedlist []int64

	//	How many bits can we glean directly from the input i.e. if n is factored, how many times does 2 show up?
	var directbits uint = (uint)(Prime_factors2(n))
	var paritybitmask int32 = (int32)(math.Pow(2, (float64)(directbits)) - 1) // so for 5 bits wil give 0b11111
	logger.Print("Direct bits: ", directbits)
	logger.Print("Parity bitmask: ", paritybitmask)
	var toTestbits = make([]int32, len(values))

	for idx, test := range toTest {
		toTestbits[idx] = (int32)(test) & paritybitmask
	}
	logger.Print("toTest ", toTest)
	logger.Print("toTestbitsvalues ", toTestbits)
	//	48 bits seed, low 16 (0-15) never shown, bit 16 also not output (see nextInt(n) - uses next(31))
	//	so directbits are from bit 17 to bit 17+directbits.  Brute them, and check parity
	var nomatch bool = false
	var maxtest = (missing + 1) * (len(toTest) - 1)
	logger.Print("maxtest ", maxtest)
	for lowseedbits := 0; lowseedbits < int(math.Pow(2, 17)); lowseedbits++ {
		var org int64 = (int64)((int64)(toTest[0])<<17) | (int64)(lowseedbits)
		result.seed = org
		nomatch = false

		var index, lastFound int = 1, 1
		for j := 1; j <= maxtest; j++ {
			//logger.Print("j ", j)
			var candidate int32 = result.NextInt(n) & paritybitmask
			//logger.Print("candidate ", candidate)
			if (n & -n) != n { // ignore powers of 2

				if candidate == (int32)(toTestbits[index]) {
					lastFound = j
					index++
					//logger.Print("candidate ", candidate)
				} else if j > lastFound+missing {
					nomatch = true
					//logger.Print("We tested all before it broke...")
					break
				}
			}
			if index == len(toTest) {
				break
			}
		}
		if !nomatch {
			logger.Printf("found lower %d bits %d", 17+directbits, org)
			result.seed = org
			seedlist = append(seedlist, org)
		}
	}

	logger.Printf("%d seeds to test", len(seedlist))

	for _, testseed := range seedlist {
		logger.Print("testing seed ", testseed)
		result.seed = testseed
		// Now for the upper bits
		var foundbits = 17 + directbits
		var uppermask = (int64)(math.Pow(2, (float64)(foundbits))) - 1 // so for 22 bits wil give 0x7fffff
		var tseed = result.seed & uppermask
		var uppercand int64 = 0
		for uppercand = 0; uppercand < (int64)(math.Pow(2, (float64)(48-foundbits))); uppercand++ {

			result.seed = uppercand<<foundbits | tseed
			nomatch = false
			var index, lastFound int = 1, 0 // was 1,1 but then false positives
			for j := 1; j <= maxtest; j++ {
				candidate := result.NextInt(n)
				//logger.Print("candidate ", candidate)
				if candidate == (int32)(toTest[index]) {
					lastFound = j
					index++
				} else if j > lastFound+missing {
					nomatch = true
					break
				}
				if index == len(toTest) {
					break
				}
			}
			if !nomatch {
				result.seed = uppercand<<foundbits | tseed

				logger.Printf("found! %d", result.seed)
				return result
			}

		}
	}
	logger.Print("Not found :-(")
	return result
}

func crackNextIntnDecr(values []string, missing int, n int32) lcg {
	//Tries to crack nextInt(n) given at least 2 values from the PRNG, no missing values, n even but not a power of 2, decreasing by 1 every time"""
	var result = NewLCG()
	if len(values) < 2 {
		return result
	}
	var toTest = make([]int, len(values))
	var err error
	for i := range values {
		toTest[i], err = strconv.Atoi(strings.TrimSpace(values[i]))
		if err != nil {
			panic(err)
		}
	}
	var seedlist []int64

	//	How many bits can we glean directly from the input i.e. if n is factored, how many times does 2 show up?
	var directbits uint = (uint)(Prime_factors2(n))
	var paritybitmask []int32 = make([]int32, len(values))
	for i := 0; i < len(values); i++ {
		paritybitmask[i] = (int32)(math.Pow(2, (float64)((uint)(Prime_factors2(n-(int32)(i))))) - 1) // so for 5 bits wil give 0b11111
		paritybitmask[i] = paritybitmask[i] & paritybitmask[0]                                       // but never more bits than what we have ... debugged... :-)
	}
	logger.Print("Direct bits: ", directbits)
	logger.Print("Parity bit mask: ", paritybitmask)
	var toTestbits = make([]int32, len(values))

	for idx, test := range toTest {
		toTestbits[idx] = (int32)(test) & paritybitmask[idx]
	}
	logger.Print("toTest ", toTest)
	logger.Print("toTestbits converted to ints ", toTestbits)
	//	48 bits seed, low 16 (0-15) never shown, bit 16 also not output (see nextInt(n) - uses next(31))
	//	so directbits are from bit 17 to bit 17+directbits.  Brute them, and check parity
	var nomatch bool = false
	var maxtest = len(toTest) - 1
	var wrongcount int = 0
	logger.Print("maxtest ", maxtest)
	for lowseedbits := 0; lowseedbits < int(math.Pow(2, 17)); lowseedbits++ {
		var org int64 = (int64)((int64)(toTest[0])<<17) | (int64)(lowseedbits)
		result.seed = org
		nomatch = false
		wrongcount = 0
		for j := 1; j <= maxtest; j++ { //how many times do we call getnext before running out of values to test?
			var testn int32 = n - (int32)(j)
			var rcandidate int32 = result.NextInt(testn)
			var candidate int32 = rcandidate & paritybitmask[j]
			//			logger.Printf("j: %d, candidate: %d , bitmask: %d", j, candidate, toTestbits[j])			//logger.Print("candidate ", candidate)
			//	if 682730 == org { logger.Printf("nomatch wrongcount %d, candidate %d, raw candidate %d testbits %d j %d", wrongcount, candidate, rcandidate, toTestbits[j], j)
			//	}
			if ((testn & -testn) != testn) && (candidate != (int32)(toTestbits[j])) { // ignore powers of 2
				wrongcount++
				if wrongcount > missing {
					nomatch = true
					break
				}
			}
		}
		if !nomatch {
			//logger.Printf("found lower %d bits %d", 17+directbits, org)
			result.seed = org
			seedlist = append(seedlist, org)
		}
	}
	logger.Printf("%d seeds to test", len(seedlist))

	for _, testseed := range seedlist {
		//var newresult = testSeed(testseed, directbits, maxtest, n, toTest)
		//if newresult.seed != 0 {
		//	return newresult
		//}
		wg.Add(1)
		go testSeed(testseed, directbits, maxtest, n, toTest, &(result.seed), missing)
	}

	result.seed = 0
	logger.Print("Waiting To Finish")
	wg.Wait()
	if result.seed == 0 {
		logger.Print("Not found :-(")
	}
	return result
}

func testSeed(seed int64, directbits uint, maxtest int, n int32, toTest []int, oldseed *int64, missing int) lcg {
	// make sure the done is called even if something goes wrong in the task by deferring
	defer wg.Done()
	//logger.Print("testing seed ", seed)
	var result = NewLCG()
	result.seed = seed
	updateSeed := func() {
		*oldseed = result.seed
	}
	var nomatch bool = false
	var wrongcount int = 0
	// Now for the upper bits
	var foundbits = 17 + directbits
	var uppermask = (int64)(math.Pow(2, (float64)(foundbits))) - 1 // so for 22 bits wil give 0x7fffff
	var tseed = result.seed & uppermask
	var uppercand int64 = 0
	for uppercand = 0; uppercand < (int64)(math.Pow(2, (float64)(48-foundbits))); uppercand++ {

		result.seed = uppercand<<foundbits | tseed
		nomatch = false
		wrongcount = 0
		for j := 1; j <= maxtest; j++ { //how many times do we call getnext before running out of values to test?
			var candidate int32 = result.NextInt(n - (int32)(j))
			//logger.Print("candidate ", candidate)
			if candidate != (int32)(toTest[j]) {
				wrongcount++
				if wrongcount > missing {
					nomatch = true
					break
				}
			}
		}
		if !nomatch {
			result.seed = uppercand<<foundbits | tseed
			once.Do(updateSeed)
			logger.Printf("found! lowerbits %d seed %d", seed, result.seed)
			return result
		}

	}

	result.seed = 0
	logger.Print("completed testing seed ", seed)
	return result
}

func main() {
	algorithmPtr := flag.String("algorithm", "LCG", "LCG for java.util.rand. Currently available: LCG")
	methodPtr := flag.String("method", "nextInt", "The method to use.  Currently available: nextInt,nextIntn,nextIntnDecr")
	probNPtr := flag.Int("probn", 0, "For nextIntn, the value of n")
	missingPtr := flag.Int("missing", 0, "up to how many values can be missing (intn) or incorrect(intndecr)")
	nextPtr := flag.Int("next", 1, "how many next values to print, if found")
	var verbosePtr = flag.Bool("verbose", false, "increase output verbosity")
	var valuesPtr = flag.String("values", "0", "the values to analyse, comma separated")
	var crackPtr = flag.Bool("crack", true, "crack")

	flag.Parse()
	if *verbosePtr {
		logger.SetOutput(os.Stdout)
	}
	logger.Print(*algorithmPtr)
	logger.Print(*valuesPtr)

	var values []string = strings.Split(*valuesPtr, ",")
	var sanity int
	var err error
	sanity, err = strconv.Atoi(strings.TrimSpace(values[1]))
	if err != nil {
		panic(err)
	}
	var generator lcg

	if *algorithmPtr == "LCG" {

		generator = NewLCG()
		generator.SetSeed(782049905)
	}

	var seedlist []int64
	if (*methodPtr == "nextInt") && *crackPtr {
		seedlist = crackNextInt(values, *missingPtr)
	}
	if (*methodPtr == "nextIntn") && *crackPtr {
		generator = crackNextIntn(values, *missingPtr, (int32)(*probNPtr))
	}
	if (*methodPtr == "nextIntnDecr") && *crackPtr {
		generator = crackNextIntnDecr(values, *missingPtr, (int32)(*probNPtr))
	}
	if seedlist != nil && *nextPtr > 0 {
		// print nextInt results
		fmt.Printf("\u001b[32mSuccess!\u001b[0m\n")
		if *methodPtr == "nextInt" {
			for _, seedFound := range seedlist {
				generator.seed = seedFound
				fmt.Printf("seed: %d. The next %d values after %s are:\n", generator.seed, *nextPtr, values[0])
				for i := 0; i < *nextPtr; i++ {
					fmt.Println(generator.NextInt32())
				}
			}
		}
	} else {
		//nextintn and nextintndecr results
		fmt.Printf("\u001b[32mSuccess!\u001b[0m\n")
		fmt.Printf("seed: %d. The next %d values after %s are:\n", generator.seed, *nextPtr, values[0])
		for i := 0; i < *nextPtr; i++ {
			if *methodPtr == "nextIntn" {
				//nextintn results
				nextValue := generator.NextInt((int32)(*probNPtr))
				if (i == 0) && (nextValue != (int32)(sanity)) {
					fmt.Printf("\u001b[33mSanity failed, rollover or missing values? %d :-(\n\u001b[0m", sanity)
					//break
				}
				fmt.Println(nextValue)
			} else {
				//nextintndecr results
				nextValue := generator.NextInt((int32)(*probNPtr - i - 1))
				if (i == 0) && (nextValue != (int32)(sanity)) {
					fmt.Printf("\u001b[33mSanity failed, rollover or incorrect values? %d :-(\n\u001b[0m", sanity)
					//break
				}
				fmt.Printf("nextint(%d): %d, seed after: %d\n", (*probNPtr - i - 1), nextValue, generator.seed)
			}
		}
	}
}

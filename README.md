# randcrack

`randcrack` is a tool used to "crack" a sequence of pseudo random numbers and predict the next terms in the list.

It currently only supports the LCG generator as used by `java.util.Random` [JavaDocs](https://docs.oracle.com/javase/9/docs/api/java/util/Random.html)

There are two versions, `randcrack_st` (single threaded) and `randcrack_mt` which is multi-threaded and the one to use if you want to max out all your CPU cores.

## method
The process to crack a LCG generator is documented [here](https://jazzy.id.au/2010/09/20/cracking_random_number_generators_part_1.html). What this tool adds is the ability to crack nextInt(n), as well as cases where the subsequent calls to nextInt(n) has a decrementing `n` as used for the [Fisher-Yates shuffle](https://en.wikipedia.org/wiki/Fisher%E2%80%93Yates_shuffle). Missing values can also be handled.

## writeup
See the PDF file for writeup and explantation.

## installation
`go get https://github/com/fransla/randcrack` should do the trick. Alternatively, binaries are located in the `bin` folder.

## targets and demos
Vulnerable Java samples are included in the `goats` folder, and (mostly) match the [asciinema](https://asciinema.org/) demos in the `demos` folder.

## usage
Full usage instructions will be put here, until then, see the demo section
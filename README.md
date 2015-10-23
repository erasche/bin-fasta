# Knapsack Problem Variant: Build Optimal Length Fasta Sequences from Ligands

Given, say, 150 different fasta sequences, we need to find an optimal way to
bin them where each bin is around 2kb.

Thus, we find ourselves with a variant of the Knapsack problem. The code for
this tool was heavily borrowed from
https://github.com/handcraftsman/GeneticGo's "samples/ukp/rosetta.go" file.

Normally in UKP, you're trying to find the maximal value of a single bin, given
some contraints like weight/volume and some motivators like "value".

We abuse the heck out of this (in order to have to write less new code) by
writing a fitness function which:

- treats "count" (number of objects per bin in UKP) as the bin number
- biasing heavily towards low bin numbers (to force multiple items into bins together)
- biasing against bins which are far from our target size (similar to weight/volume)
- biasing heavily against solutions which do not use all of the available resources (i.e. sequences)


## Usage

```console
$ go run bin_fasta.go -lengthTable test.tsv -targetLength=2000 > out.tsv
```

This will take the features defined in `test.tsv` and their lengths in column
2, and try to group them such that each grouping is approximately 2kb. See
`-help` for options.

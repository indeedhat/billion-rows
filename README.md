# Billion row challenge
This repo contains my attempts at the billion row challenge starting from a very basic and expectedly slow (iirc about 2-2.5 minutes)
single threaded approach and being iterated on each commit, changes on each commit are documented in the readme for said commit

## last changes
- perform string conversion on chunk before passing to chunk channel to optimize memcopy
- fix bug where i was creting a buffer 1 byte too large to deal with chunk overflow

## current implementation measurement
```console
real    0m6.805s
user    1m39.661s
sys     0m2.179s
```

## previous implementation measurement
```console
real    0m7.620s
user    1m34.781s
sys     0m4.370s
```

## Baseline measurements
```console
real    3m49.808s
user    3m38.623s
sys     0m10.442s
```

## output format
%s=%0.1f,%0.1f,%0.1f, ...

## Original challenge
[Can be found here](https://github.com/gunnarmorling/1brc)

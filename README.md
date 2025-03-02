# Signal Party!

![Signal Party](./logo.png)

[![Go Reference](https://pkg.go.dev/badge/github.com/delaneyj/signalparty.svg)](https://pkg.go.dev/github.com/delaneyj/signalparty)

## Implementations

* [Alien](https://github.com/stackblitz/alien-signals) is currently the fastest mainstream signal library for JavaScript.
* [Reactievly](https://github.com/milomg/reactively) is another fast signal library for JavaScript but has been overshadowed by Alien.
* Dumbdumb is the simplest approach to signal propagation.  It's using the core mechanisms of how most spreadsheet engines work.
* 🚀 is a combination of core ideas of [zignals](https://github.com/jmstevers/zignals/blob/main/src/effect.zig) combined with explicit code gen

> [!WARNING]
> Dumbdumb and 🚀 are the only ones thread safe!  This is actually togglable in codegen but the numbers are good enough it's left on in the benchmarks as it's better in a real world sense. This is not about distributing the workload, more about access safety

## Benchmarks

```bash
+----------------------------------------------------------------------------------------------+
| Alien Signals                                                                                |
+------------------------+-------------+-------------+-------------+-------------+-------------+
| BENCHMARK              |         AVG |         MIN |         P75 |         P99 |         MAX |
+------------------------+-------------+-------------+-------------+-------------+-------------+
| propagate: 1 * 1       |       202ns |       170ns |       190ns |       371ns |     2.084µs |
| propagate: 1 * 10      |     1.174µs |       942ns |     1.253µs |     1.383µs |     1.654µs |
| propagate: 1 * 100     |     7.004µs |     6.622µs |     6.673µs |     8.436µs |    17.474µs |
| propagate: 1 * 1000    |    65.993µs |    65.045µs |    65.286µs |    89.032µs |    92.638µs |
| propagate: 10 * 1      |     1.217µs |     1.202µs |     1.213µs |     1.323µs |     1.653µs |
| propagate: 10 * 10     |     7.474µs |     7.374µs |     7.394µs |     7.915µs |     15.12µs |
| propagate: 10 * 100    |    71.274µs |    63.662µs |    66.578µs |   111.164µs |   117.817µs |
| propagate: 10 * 1000   |   512.438µs |   448.674µs |   497.348µs |   961.271µs |  1.244056ms |
| propagate: 100 * 1     |     8.203µs |     7.684µs |     7.745µs |    12.744µs |    13.396µs |
| propagate: 100 * 10    |    57.081µs |    48.544µs |    60.066µs |    87.429µs |    90.013µs |
| propagate: 100 * 100   |   475.861µs |   454.645µs |    477.35µs |   617.058µs |   625.515µs |
| propagate: 100 * 1000  |  4.845889ms |  4.645799ms |  4.878147ms |  5.516667ms |  5.635034ms |
| propagate: 1000 * 1    |    72.038µs |    70.876µs |     72.46µs |    79.814µs |    80.686µs |
| propagate: 1000 * 10   |    520.44µs |   486.667µs |   522.837µs |   613.511µs |   620.695µs |
| propagate: 1000 * 100  |  5.474449ms |  5.026042ms |  5.651396ms |   6.39089ms |  6.723992ms |
| propagate: 1000 * 1000 | 53.128055ms | 51.350893ms | 53.526383ms | 57.249605ms | 58.541503ms |
+------------------------+-------------+-------------+-------------+-------------+-------------+
+---------------------------------------------------------------------------------------------+
| dumbdumb Signals                                                                            |
+------------------------+-------------+------------+-------------+-------------+-------------+
| BENCHMARK              |         AVG |        MIN |         P75 |         P99 |         MAX |
+------------------------+-------------+------------+-------------+-------------+-------------+
| propagate: 1 * 1       |        72ns |       60ns |        70ns |       120ns |       340ns |
| propagate: 1 * 10      |       269ns |      260ns |       271ns |       331ns |       411ns |
| propagate: 1 * 100     |     2.017µs |    2.004µs |     2.023µs |     2.054µs |     2.084µs |
| propagate: 1 * 1000    |    41.207µs |   38.825µs |     41.38µs |    54.295µs |    59.605µs |
| propagate: 10 * 1      |       417ns |      410ns |       421ns |       541ns |       631ns |
| propagate: 10 * 10     |     2.638µs |    2.515µs |     2.535µs |     4.208µs |     8.005µs |
| propagate: 10 * 100    |     20.89µs |     20.6µs |     20.69µs |    28.495µs |    28.655µs |
| propagate: 10 * 1000   |   486.566µs |  398.898µs |    451.94µs |  1.262101ms |   1.32389ms |
| propagate: 100 * 1     |     5.241µs |    5.159µs |      5.18µs |     5.992µs |     8.236µs |
| propagate: 100 * 10    |    26.278µs |    25.82µs |     26.02µs |    32.833µs |    35.148µs |
| propagate: 100 * 100   |   251.162µs |  237.037µs |   255.302µs |   312.963µs |   327.521µs |
| propagate: 100 * 1000  |  5.057141ms | 4.123353ms |  5.551614ms |    6.7335ms |  6.920631ms |
| propagate: 1000 * 1    |     56.86µs |   56.198µs |    56.639µs |    61.018µs |    65.747µs |
| propagate: 1000 * 10   |   297.857µs |  243.439µs |   298.385µs |   570.008µs |    627.87µs |
| propagate: 1000 * 100  |  2.392515ms |  2.28427ms |  2.404722ms |   2.89187ms |  2.971233ms |
| propagate: 1000 * 1000 | 53.532082ms | 47.91253ms | 55.689851ms | 59.346284ms | 61.098078ms |
+------------------------+-------------+------------+-------------+-------------+-------------+
+----------------------------------------------------------------------------------------------+
| 🚀 Signals                                                                                   |
+------------------------+-------------+-------------+-------------+-------------+-------------+
| BENCHMARK              |         AVG |         MIN |         P75 |         P99 |         MAX |
+------------------------+-------------+-------------+-------------+-------------+-------------+
| propagate: 1 * 1       |        47ns |        30ns |        40ns |       130ns |       732ns |
| propagate: 1 * 10      |       236ns |       230ns |       231ns |       350ns |       361ns |
| propagate: 1 * 100     |     2.834µs |     2.194µs |     2.916µs |     3.026µs |     3.758µs |
| propagate: 1 * 1000    |    37.778µs |    30.339µs |    38.354µs |    53.673µs |    72.619µs |
| propagate: 10 * 1      |       191ns |       190ns |       191ns |       210ns |       261ns |
| propagate: 10 * 10     |     1.973µs |     1.953µs |     1.974µs |     2.014µs |     2.185µs |
| propagate: 10 * 100    |    27.286µs |    25.309µs |    27.382µs |    34.787µs |    39.316µs |
| propagate: 10 * 1000   |    390.15µs |    368.73µs |   394.389µs |   440.208µs |   453.323µs |
| propagate: 100 * 1     |     1.742µs |     1.733µs |     1.743µs |     1.773µs |     2.004µs |
| propagate: 100 * 10    |     22.21µs |    21.962µs |    22.142µs |    26.381µs |    26.521µs |
| propagate: 100 * 100   |   280.615µs |   270.842µs |   281.422µs |   303.926µs |   323.092µs |
| propagate: 100 * 1000  |  4.135867ms |   3.69602ms |  4.207275ms |  5.153647ms |  6.389858ms |
| propagate: 1000 * 1    |    15.553µs |    14.909µs |    15.079µs |    25.278µs |    26.251µs |
| propagate: 1000 * 10   |   248.248µs |   214.964µs |   234.361µs |   501.937µs |   706.982µs |
| propagate: 1000 * 100  |  2.984119ms |  2.800955ms |  3.061107ms |  3.370582ms |  3.451528ms |
| propagate: 1000 * 1000 | 44.106551ms | 42.115223ms | 44.546296ms | 49.121499ms | 49.432287ms |
+------------------------+-------------+-------------+-------------+-------------+-------------+
```
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
+---------------------------------------------------------------------------------------------+
| Alien Signals                                                                               |
+------------------------+-------------+-------------+------------+-------------+-------------+
| BENCHMARK              |         AVG |         MIN |        P75 |         P99 |         MAX |
+------------------------+-------------+-------------+------------+-------------+-------------+
| propagate: 1 * 1       |       222ns |       180ns |      211ns |       411ns |     1.633µs |
| propagate: 1 * 10      |     1.089µs |       671ns |    1.533µs |     1.754µs |     1.824µs |
| propagate: 1 * 100     |     5.572µs |     4.619µs |    5.661µs |     10.27µs |    11.142µs |
| propagate: 1 * 1000    |    49.623µs |    45.778µs |    51.83µs |    69.734µs |    70.185µs |
| propagate: 10 * 1      |       993ns |       851ns |      872ns |     1.563µs |     1.583µs |
| propagate: 10 * 10     |     5.404µs |      5.13µs |      5.2µs |     9.048µs |     9.548µs |
| propagate: 10 * 100    |    55.119µs |     46.89µs |   59.124µs |    82.088µs |    84.833µs |
| propagate: 10 * 1000   |   572.821µs |   486.267µs |  591.349µs |   769.512µs |   780.203µs |
| propagate: 100 * 1     |     9.203µs |     7.765µs |    7.996µs |    24.718µs |    39.566µs |
| propagate: 100 * 10    |    54.063µs |    49.686µs |   61.078µs |    70.976µs |    71.989µs |
| propagate: 100 * 100   |    543.76µs |   485.716µs |   545.11µs |   763.641µs |   784.461µs |
| propagate: 100 * 1000  |  5.651017ms |  5.434287ms | 5.718484ms |  6.070253ms |  6.073218ms |
| propagate: 1000 * 1    |        83µs |    76.748µs |   88.431µs |   102.607µs |   104.231µs |
| propagate: 1000 * 10   |   578.923µs |   531.303µs |  584.987µs |   683.466µs |   730.086µs |
| propagate: 1000 * 100  |   6.49256ms |  6.070813ms | 6.602768ms |  6.956339ms |  7.016996ms |
| propagate: 1000 * 1000 | 61.534548ms | 55.628069ms |  64.4648ms | 75.700571ms | 75.889445ms |
+------------------------+-------------+-------------+------------+-------------+-------------+
+---------------------------------------------------------------------------------------------+
| 🚀 Signals                                                                                  |
+------------------------+-------------+------------+-------------+-------------+-------------+
| BENCHMARK              |         AVG |        MIN |         P75 |         P99 |         MAX |
+------------------------+-------------+------------+-------------+-------------+-------------+
| propagate: 1 * 1       |        45ns |       30ns |        40ns |        90ns |       652ns |
| propagate: 1 * 10      |       215ns |      210ns |       220ns |       271ns |       320ns |
| propagate: 1 * 100     |     2.424µs |    2.384µs |     2.425µs |     2.545µs |     3.397µs |
| propagate: 1 * 1000    |    40.016µs |   34.166µs |    40.357µs |    69.023µs |    71.718µs |
| propagate: 10 * 1      |       181ns |      170ns |       181ns |       191ns |       220ns |
| propagate: 10 * 10     |      1.89µs |    1.654µs |     2.074µs |     2.585µs |     2.695µs |
| propagate: 10 * 100    |    24.875µs |   23.806µs |    24.337µs |    34.156µs |     46.72µs |
| propagate: 10 * 1000   |   433.334µs |  396.774µs |   445.479µs |    524.58µs |   524.761µs |
| propagate: 100 * 1     |     1.515µs |    1.493µs |     1.522µs |     1.553µs |     1.624µs |
| propagate: 100 * 10    |    18.484µs |   17.553µs |    17.965µs |    24.256µs |    25.319µs |
| propagate: 100 * 100   |   266.185µs |  252.456µs |   273.216µs |   303.845µs |   311.309µs |
| propagate: 100 * 1000  |  4.535147ms | 4.213165ms |  4.626762ms |  5.294297ms |  5.382668ms |
| propagate: 1000 * 1    |    15.101µs |   14.628µs |    14.738µs |    22.704µs |    32.313µs |
| propagate: 1000 * 10   |   186.796µs |  177.131µs |   191.369µs |   213.601µs |   222.348µs |
| propagate: 1000 * 100  |  3.140367ms |  2.96454ms |  3.241504ms |  3.506755ms |  3.603351ms |
| propagate: 1000 * 1000 | 42.979905ms | 39.66941ms | 44.120654ms | 50.342507ms | 54.134532ms |
+------------------------+-------------+------------+-------------+-------------+-------------+
+----------------------------------------------------------------------------------------------+
| dumbdumb Signals                                                                             |
+------------------------+-------------+-------------+-------------+-------------+-------------+
| BENCHMARK              |         AVG |         MIN |         P75 |         P99 |         MAX |
+------------------------+-------------+-------------+-------------+-------------+-------------+
| propagate: 1 * 1       |        73ns |        70ns |        70ns |       140ns |       290ns |
| propagate: 1 * 10      |       281ns |       270ns |       281ns |       330ns |       401ns |
| propagate: 1 * 100     |     2.138µs |     2.124µs |     2.144µs |     2.214µs |     2.215µs |
| propagate: 1 * 1000    |    45.099µs |    40.828µs |    45.117µs |    63.833µs |    72.911µs |
| propagate: 10 * 1      |       430ns |       420ns |       431ns |       531ns |       651ns |
| propagate: 10 * 10     |     2.619µs |     2.605µs |     2.616µs |     2.705µs |     2.785µs |
| propagate: 10 * 100    |    22.036µs |    21.071µs |     21.15µs |    33.955µs |    34.166µs |
| propagate: 10 * 1000   |   470.663µs |   433.124µs |   477.771µs |    498.48µs |   502.117µs |
| propagate: 100 * 1     |     5.645µs |      5.49µs |     5.541µs |     7.204µs |    11.492µs |
| propagate: 100 * 10    |    30.072µs |    29.717µs |    29.818µs |    34.085µs |    38.765µs |
| propagate: 100 * 100   |    246.87µs |   242.537µs |   248.538µs |   266.023µs |   267.725µs |
| propagate: 100 * 1000  |  5.623111ms |  4.544303ms |  5.998424ms |   8.58181ms |  9.340281ms |
| propagate: 1000 * 1    |    58.768µs |     56.96µs |    58.353µs |    71.187µs |    78.932µs |
| propagate: 1000 * 10   |   284.962µs |   275.752µs |    290.87µs |   302.333µs |    306.34µs |
| propagate: 1000 * 100  |   2.35285ms |  2.194437ms |  2.398129ms |  2.756599ms |   2.82392ms |
| propagate: 1000 * 1000 | 52.749424ms | 48.032408ms | 54.721442ms | 56.868538ms | 58.178531ms |
+------------------------+-------------+-------------+-------------+-------------+-------------+
```
# Download objective comparison

All four grids contain 21 workloads. Endpoint = minimum-download frontier endpoint for searches, or selected greedy plan.

| Experiment | Changed endpoint costs | Changed displayed rows | Changed intermediate frontiers |
|---|---:|---:|---:|
| greedy-costs | 0/21 | — | — |
| application-costs | 0/21 | — | — |
| tradeoff-costs | 0/21 | 1 | 21 |
| dp-costs | 0/21 | 1 | 21 |

Display changes:

tradeoff-costs:
| Version | Database | Record | Reference N/K · D/B | Lower-storage option | Minimum download option |
|---|---|---|---:|---:|---:|
| Before | 20 GB | 10 KB | 4.438× · 131,199× | 8.652× · 6,552× (65.522 MB) ★ | 8.913× · 5,374× (53.742 MB) |
| After | 20 GB | 10 KB | 4.438× · 131,199× | 8.684× · 5,884× (58.842 MB) ★ | 8.913× · 5,374× (53.742 MB) |

dp-costs:
| Version | Database | Record | Reference N/K · D/B | Lower-storage option | Minimum download option |
|---|---|---|---:|---:|---:|
| Before | 20 GB | 10 KB | 4.438× · 131,199× | 8.652× · 6,552× (65.522 MB) ★ | 8.913× · 5,374× (53.742 MB) |
| After | 20 GB | 10 KB | 4.438× · 131,199× | 8.684× · 5,884× (58.842 MB) ★ | 8.913× · 5,374× (53.742 MB) |

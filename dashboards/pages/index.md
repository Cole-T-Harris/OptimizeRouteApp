---
title: Commutes Optimizer Dashboard
---

```sql users
select name 
from users 
group by name
```

<Dropdown data={users} name=user_dropdown value=name title="Select your User Name">
</Dropdown>

<Dropdown name=to_work title="Commuting to Work?" defaultValue="true">
  <DropdownOption value="true" valueLabel="Yes"/>
  <DropdownOption value="false" valueLabel="No"/>
</Dropdown>

```sql commutes_chart
WITH formatted_commutes AS (
  SELECT
    strftime(adjusted_query_time, '%H:%M') as formatted_time,
    duration,
    day_of_week,
    route,
    to_work,
    Routes.active AS Routes__active,
    Routes.id AS route_id,
    Users.name AS Users__name
  FROM commutes
  LEFT JOIN routes AS Routes ON commutes.route = Routes.id
  LEFT JOIN users AS Users ON commutes.user_id = Users.id
)
SELECT
  formatted_time as query_time,
  AVG(duration) / 60 AS avg,
  day_of_week,
  Users__name,
  route,
  Routes__active,
  to_work
FROM formatted_commutes
WHERE
  Users__name = '${inputs.user_dropdown.value}'
  AND to_work = '${inputs.to_work.value}'
GROUP BY
  formatted_time,
  day_of_week,
  Users__name,
  route,
  Routes__active,
  to_work
ORDER BY
  formatted_time ASC
```

```sql max_commute_time
SELECT 
  duration / 60 as max_duration_minutes, 
  adjusted_query_time,
  day_of_week
FROM commutes
LEFT JOIN users AS Users ON commutes.user_id = Users.id
WHERE
  Users.name = '${inputs.user_dropdown.value}'
  AND to_work = '${inputs.to_work.value}'
ORDER BY duration DESC
LIMIT 1
```

<LineChart
  data={commutes_chart}
  title="Commute Average By Day of Week"
  x=query_time
  y=avg
  yAxisTitle="Commute Time (Minutes)"
  xAxisTitle="Commute Leaving Time"
  xFmt="H:MM:SS AM/PM"
  series=day_of_week
  sort=false
  yTickMarks=true
  yScale=true
  chartAreaHeight=360
  downloadableData=false
>
  <ReferenceLine data="{max_commute_time}" y="max_duration_minutes" label="Max Duration Time" />
</LineChart>

```sql avg_commute_time_per_day
WITH overall_avg AS (
  SELECT AVG(duration) / 60 AS overall_avg_time
  FROM commutes
  LEFT JOIN users AS Users ON commutes.user_id = Users.id
  WHERE
    Users.name = '${inputs.user_dropdown.value}'
    AND to_work = '${inputs.to_work.value}'
),
daily_avg AS (
  SELECT AVG(duration) / 60 AS avg_time, day_of_week
  FROM commutes
  LEFT JOIN users AS Users ON commutes.user_id = Users.id
  WHERE
    Users.name = '${inputs.user_dropdown.value}'
    AND to_work = '${inputs.to_work.value}'
  GROUP BY day_of_week
)
SELECT 
  daily_avg.day_of_week,
  daily_avg.avg_time,
  overall_avg.overall_avg_time,
  ((daily_avg.avg_time - overall_avg.overall_avg_time) / overall_avg.overall_avg_time) AS percent_diff
FROM daily_avg
CROSS JOIN overall_avg
```

{#each avg_commute_time_per_day as avg_stat}
  <BigValue
    data={avg_stat}
    value=avg_time
    fmt='0.0 "min"'
    title={avg_stat.day_of_week}
    comparison=percent_diff
    comparisonTitle="vs. Avg"
    comparisonFmt=pct1
    downIsGood=true
  />
{/each}

## What's Next?
- [Connect your data sources](settings)
- Edit/add markdown files in the `pages` folder
- Deploy your project with [Evidence Cloud](https://evidence.dev/cloud)

## Get Support
- Message us on [Slack](https://slack.evidence.dev/)
- Read the [Docs](https://docs.evidence.dev/)
- Open an issue on [Github](https://github.com/evidence-dev/evidence)

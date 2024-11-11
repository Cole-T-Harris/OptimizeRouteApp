---
title: Commutes Optimizer Dashboard
---

<LastRefreshed prefix="Data last updated"/>

```sql users
select name 
from users 
group by name
```

<Dropdown data={users} name=user_dropdown value=name title="Select your User Name">
</Dropdown>

```sql routes
SELECT routes.id, start_address, end_address, CONCAT(start_address, ' - ', end_address) as total_route
FROM routes
LEFT JOIN users on users.id = routes.user_id
WHERE users.name = '${inputs.user_dropdown.value}'
```

<Dropdown data={routes} name=route_dropdown value=id label=total_route title="Select your Route">
</Dropdown>

<Checkbox
  title="Commuting to Work?"
  name=to_work
  defaultValue="true"
/>

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
  AND to_work = '${inputs.to_work}'
  AND route_id = '${inputs.route_dropdown.value}'
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
LEFT JOIN routes AS Routes ON commutes.route = Routes.id
WHERE
  Users.name = '${inputs.user_dropdown.value}'
  AND to_work = '${inputs.to_work}'
  AND routes.id = '${inputs.route_dropdown.value}'
ORDER BY duration DESC
LIMIT 1
```

### Commute Average By Day of Week

<LineChart
  data={commutes_chart}
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
  LEFT JOIN routes AS Routes ON commutes.route = Routes.id
  WHERE
    Users.name = '${inputs.user_dropdown.value}'
    AND to_work = '${inputs.to_work}'
    AND routes.id = '${inputs.route_dropdown.value}'
),
daily_avg AS (
  SELECT AVG(duration) / 60 AS avg_time, day_of_week
  FROM commutes
  LEFT JOIN users AS Users ON commutes.user_id = Users.id
  LEFT JOIN routes AS Routes ON commutes.route = Routes.id
  WHERE
    Users.name = '${inputs.user_dropdown.value}'
    AND to_work = '${inputs.to_work}'
    AND routes.id = '${inputs.route_dropdown.value}'
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
### Average Commute Times

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

### Approximate Optimal Departure Times 

<Details title="Calculation Explanation">

  I will not hide that the below process is AI generated, if there are errors, please report.

  ## Inflection Point Calculation Process

  Below outlines the process used to identify inflection points in commute time data. These inflection points mark the times at which the commute duration starts to increase significantly, helping you determine the latest optimal departure time.

  ## Step-by-Step Explanation of Calculations

  ### 1. Data Preparation

  - `formatted_commutes`: Prepares the raw commute data by formatting the time and joining additional information (such as user and route details) for filtering and grouping.
  - `average_commutes`: Calculates the average commute duration (`avg_duration`) at each `query_time`, grouped by day of the week, user, and whether you are commuting to or from work. This grouping allows us to analyze changes in average commute duration over time.

  ### 2. First Derivative Calculation

  The **first derivative** represents the rate of change in commute duration over time. 

  ```sql
  duration_change = avg_duration - LAG(avg_duration)
  ```
  - Interpretation: A high `duration_change` value between two adjacent times indicates a large shift in commute time. This captures when the commute duration starts to rise or fall sharply.

  ### 3. Second Derivative Calculation

  The **second derivative** represents the rate of change of the first derivative, `duration_change`.

  ```sql
  rate_of_change = duration_change - LAG(duration_change)
  ```

  - Interpretation: A high `rate_of_change` value suggests a rapid increase or decrease in commute time. Points where this rate of change crosses a threshold can indicate inflection points, showing times of sharp acceleration in commute delays.

  ### 4. Setting a Threshold for Inflection Points

  In `threshold_values`, we calculate statistics to approximate a threshold for significant rate changes:

  - **Median of `rate_of_change`**: Provides a central value to separate typical changes from more extreme ones.
  - **Standard Deviation of `rate_of_change`**: Helps identify values that significantly exceed typical fluctuations.

Threshold for inflection points
  ```sql
  threshold = median_rate + stddev_rate
  ```
  This threshold is used to flag points where the `rate_of_change` crosses from below to above the threshold, marking these as potential inflection points.

  ### 5. Identifying Inflection Points

  - `inflection_points`: Flags rows where the rate of change crosses the threshold from below, signifying a sharp increase in commute duration.
  - `ranked_inflection_points`: Ranks inflection points by time, retaining only the top three per day of the week to limit results to the most significant changes.

  ### Final Query Output

  The final result returns up to three significant inflection points per day, sorted by time, allowing you to identify optimal departure times with minimal impact from rising commute delays.

</Details>

<!-- TODO: Fix the repetetiveness of these queries and tabs because its awful -->

```sql best_commute_time
WITH formatted_commutes AS (
  SELECT
    strftime(adjusted_query_time, '%H:%M') AS formatted_time,
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
),
average_commutes AS (
  SELECT
    formatted_time AS query_time,
    AVG(duration) / 60 AS avg_duration,
    MAX(duration) / 60 AS worst_duration,
    MIN(duration) / 60 AS best_duration,
    day_of_week,
    Users__name,
    route,
    Routes__active,
    to_work
  FROM formatted_commutes
  WHERE
    Users__name = '${inputs.user_dropdown.value}'
    AND to_work = '${inputs.to_work}'
    AND route_id = '${inputs.route_dropdown.value}'
  GROUP BY
    formatted_time,
    day_of_week,
    Users__name,
    route,
    Routes__active,
    to_work
),
first_derivative AS (
  SELECT
    query_time,
    day_of_week,
    Users__name,
    route,
    Routes__active,
    to_work,
    avg_duration,
    worst_duration,
    best_duration,
    LAG(avg_duration) OVER (PARTITION BY day_of_week ORDER BY query_time) AS previous_avg_duration,
    avg_duration - LAG(avg_duration) OVER (PARTITION BY day_of_week ORDER BY query_time) AS duration_change
  FROM average_commutes
),
second_derivative AS (
  SELECT
    query_time,
    day_of_week,
    Users__name,
    route,
    Routes__active,
    to_work,
    avg_duration,
    worst_duration,
    best_duration,
    duration_change,
    LAG(duration_change) OVER (PARTITION BY day_of_week ORDER BY query_time) AS previous_duration_change,
    duration_change - LAG(duration_change) OVER (PARTITION BY day_of_week ORDER BY query_time) AS rate_of_change
  FROM first_derivative
),
threshold_values AS (
  SELECT 
    MEDIAN(rate_of_change) AS median_rate,
    STDDEV(rate_of_change) AS stddev_rate
  FROM second_derivative
  WHERE rate_of_change IS NOT NULL
),
inflection_points AS (
  SELECT *,
    CASE 
      WHEN rate_of_change > (SELECT median_rate FROM threshold_values) + (SELECT stddev_rate FROM threshold_values)
        AND LAG(rate_of_change) OVER (PARTITION BY day_of_week ORDER BY query_time) <= (SELECT median_rate FROM threshold_values) + (SELECT stddev_rate FROM threshold_values)
      THEN 1
      ELSE 0
    END AS is_inflection_point
  FROM second_derivative
),
ranked_inflection_points AS (
  SELECT *,
    ROW_NUMBER() OVER (PARTITION BY day_of_week ORDER BY query_time ASC) AS rank
  FROM inflection_points
  WHERE is_inflection_point = 1
)
SELECT *
FROM ranked_inflection_points
WHERE rank <= 3
ORDER BY day_of_week, rank;
```

```sql generic_commutes
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
  (AVG(duration) + STDDEV(duration)) / 60 AS upper_bound,
  (AVG(duration) - STDDEV(duration)) / 60 AS lower_bound,
  day_of_week,
  Users__name,
  route,
  Routes__active,
  to_work
FROM formatted_commutes
WHERE
  Users__name = '${inputs.user_dropdown.value}'
  AND to_work = '${inputs.to_work}'
  AND route_id = '${inputs.route_dropdown.value}'
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

```sql days_of_week_query
SELECT distinct day_of_week
FROM commutes
LEFT JOIN users AS Users ON commutes.user_id = Users.id
LEFT JOIN routes AS Routes ON commutes.route = Routes.id
WHERE
  Users.name = '${inputs.user_dropdown.value}'
  AND to_work = '${inputs.to_work}'
  AND routes.id = '${inputs.route_dropdown.value}'
```

<Tabs>
  {#each days_of_week_query as days}
    <Tab label={days.day_of_week}>


        **Top 3 Optimal Departure Times**

        <DataTable data={best_commute_time.where(`day_of_week='${days.day_of_week}'`)}>
          <Column id=query_time title="Leave Time"/>
          <Column id=avg_duration title="Average Duration"/>
          <Column id=best_duration title="Best Commute Time"/>
          <Column id=worst_duration title="Worst Commute Time"/>
        </DataTable>

        <Chart 
          data={generic_commutes.where(`day_of_week='${days.day_of_week}'`)}
          x=query_time
          yAxisTitle="Commute Time (Minutes)"
          xAxisTitle="Commute Leaving Time"
          yScale=true
          sort=false
          yTickMarks=true
          emptySet=pass
          emptyMessage="Route does not contain data for this day of week"
        >
          <Line
            y=avg
          />
          <ReferenceLine 
            data={best_commute_time.where(`day_of_week='${days.day_of_week}'`)}
            x=query_time
            labelPosition=aboveStart
          />
          <Area
            y=upper_bound
            name="Upper Confidence Interval"
          />
          <Area
            y=lower_bound
            name="Lower Confidence Interval"
            fillColor=white
            fillOpacity=1
          />
        </Chart>
    </Tab>
  {/each}
</Tabs>

### Most Common Routes

```sql most_common_routes
SELECT 
  COUNT(*) as count, 
  CONCAT('https://valhalla.github.io/demos/polyline/?unescape=false&polyline6=false#', commutes.route_hash) as route_link, 
  AVG(commutes.distance) as avg_dist, 
  AVG(commutes.duration) / 60 as avg_duration
FROM commutes
LEFT JOIN users AS Users ON commutes.user_id = Users.id
LEFT JOIN routes AS Routes ON commutes.route = Routes.id
WHERE
  Users.name = '${inputs.user_dropdown.value}'
  AND to_work = '${inputs.to_work}'
  AND routes.id = '${inputs.route_dropdown.value}'
GROUP BY commutes.route_hash
ORDER by count desc
```
<DataTable 
  data={most_common_routes}
>
  <Column id=count title="Occurences" contentType=bar/>
  <Column id=avg_duration title="Average Duration"/>
  <Column id=avg_dist title="Average Distance"/>
  <Column id=route_link title="Route URL" contentType=link linkLabel="Details ->"/>
</DataTable>

### Route Information
```sql route_information
select * 
from routes
where routes.id='${inputs.route_dropdown.value}'
```
Route Status: **{route_information[0].active}**

Route Starting Address: **{route_information[0].start_address}**

Route Ending Address: **{route_information[0].end_address}**

Route Started Polling: **{route_information[0].start_date}**

{#if !route_information[0].start_date}
Route Ended Polling: **{route_information[0].end_date}**

{/if}

Route Timezone: **{route_information[0].time_zone}**
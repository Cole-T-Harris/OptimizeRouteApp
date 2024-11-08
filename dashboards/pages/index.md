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
  AND to_work = '${inputs.to_work}'
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
  WHERE
    Users.name = '${inputs.user_dropdown.value}'
    AND to_work = '${inputs.to_work}'
),
daily_avg AS (
  SELECT AVG(duration) / 60 AS avg_time, day_of_week
  FROM commutes
  LEFT JOIN users AS Users ON commutes.user_id = Users.id
  WHERE
    Users.name = '${inputs.user_dropdown.value}'
    AND to_work = '${inputs.to_work}'
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

### Optimal Departure Times 

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
    day_of_week,
    Users__name,
    route,
    Routes__active,
    to_work
  FROM formatted_commutes
  WHERE
    Users__name = '${inputs.user_dropdown.value}'
    AND to_work = '${inputs.to_work}'
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
ORDER BY day_of_week, query_time ASC;
```

<Tabs>
    <Tab label="First Tab">
        Content of the First Tab

        You can use **markdown** here too!
    </Tab>
    <Tab label="Second Tab">
        Content of the Second Tab

        Here's a [link](https://www.google.com)
    </Tab>
    <Tab label="First Tab">
        Content of the First Tab

        You can use **markdown** here too!
    </Tab>
    <Tab label="Second Tab">
        Content of the Second Tab

        Here's a [link](https://www.google.com)
    </Tab>
</Tabs>

<!-- <LineChart
  data={best_commute_time}
  title="Commute Second Derivative"
  x=query_time
  y=duration_change
  yAxisTitle="Second Derivative"
  xAxisTitle="Commute Leaving Time"
  xFmt="H:MM:SS AM/PM"
  series=day_of_week
  sort=false
  yTickMarks=true
  yScale=true
  chartAreaHeight=360
  downloadableData=false
/> -->
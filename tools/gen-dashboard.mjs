import fs from "node:fs";
const ds = { type: "prometheus", uid: "prometheus" };
let id = 0;
const t = (expr, legendFormat = "", refId = "A") => ({
  refId,
  expr,
  legendFormat,
  datasource: ds,
});
const stat = (
  title,
  expr,
  x,
  { unit = "short", color = "blue", bad = false, decimals } = {},
) => ({
  id: ++id,
  type: "stat",
  title,
  datasource: ds,
  gridPos: { x, y: 0, w: 4, h: 5 },
  targets: [t(expr)],
  fieldConfig: {
    defaults: {
      unit,
      decimals,
      color: { mode: "thresholds" },
      thresholds: {
        mode: "absolute",
        steps: bad
          ? [
              { color: "green", value: null },
              { color: "red", value: 1 },
            ]
          : [{ color, value: null }],
      },
    },
    overrides: [],
  },
  options: {
    reduceOptions: { calcs: ["lastNotNull"], fields: "", values: false },
    colorMode: "background",
    graphMode: "area",
    textMode: "value",
    justifyMode: "center",
    orientation: "auto",
  },
});
const color = (name, c) => ({
  matcher: { id: "byName", options: name },
  properties: [{ id: "color", value: { mode: "fixed", fixedColor: c } }],
});
const series = (
  title,
  targets,
  gridPos,
  {
    unit = "short",
    draw = "line",
    description,
    overrides = [],
    legend = "bottom",
  } = {},
) => ({
  id: ++id,
  type: "timeseries",
  title,
  description,
  datasource: ds,
  gridPos,
  targets,
  fieldConfig: {
    defaults: {
      unit,
      custom: {
        drawStyle: draw,
        lineWidth: 2,
        fillOpacity: draw === "bars" ? 80 : 12,
        showPoints: "never",
        stacking: { mode: draw === "bars" ? "normal" : "none" },
      },
    },
    overrides,
  },
  options: {
    legend: {
      displayMode: "table",
      placement: legend,
      calcs: ["lastNotNull", "max"],
    },
    tooltip: { mode: "multi", sort: "desc" },
  },
});
const panels = [
  stat("Всего запросов", "sum(http_requests_total)", 0),
  stat("Ошибки 4xx", 'sum(http_errors_total{class="4xx"}) or vector(0)', 4, {
    bad: true,
  }),
  stat("Ошибки 5xx", 'sum(http_errors_total{class="5xx"}) or vector(0)', 8, {
    bad: true,
  }),
  stat(
    "Успешные входы",
    'sum(auth_attempts_total{result="success"}) or vector(0)',
    12,
    { color: "green" },
  ),
  stat(
    "Неуспешные входы",
    'sum(auth_attempts_total{result="failure"}) or vector(0)',
    16,
    { bad: true },
  ),
  stat(
    "Время ответа p95 (5 мин)",
    "histogram_quantile(0.95, sum by (le) (rate(http_request_duration_seconds_bucket[5m])))",
    20,
    { unit: "s", color: "purple", decimals: 1 },
  ),
  series(
    "Запросы в секунду по маршрутам",
    [t("sum by (route) (rate(http_requests_total[1m]))", "{{route}}")],
    { x: 0, y: 5, w: 12, h: 9 },
    { unit: "reqps" },
  ),
  series(
    "Время ответа",
    [
      t(
        "histogram_quantile(0.50, sum by (le) (rate(http_request_duration_seconds_bucket[1m])))",
        "p50",
        "A",
      ),
      t(
        "histogram_quantile(0.95, sum by (le) (rate(http_request_duration_seconds_bucket[1m])))",
        "p95",
        "B",
      ),
      t(
        "histogram_quantile(0.99, sum by (le) (rate(http_request_duration_seconds_bucket[1m])))",
        "p99",
        "C",
      ),
    ],
    { x: 12, y: 5, w: 12, h: 9 },
    { unit: "s" },
  ),
  series(
    "Ошибки по классу (за минуту)",
    [t("sum by (class) (increase(http_errors_total[1m]))", "{{class}}")],
    { x: 0, y: 14, w: 8, h: 9 },
    { draw: "bars", overrides: [color("4xx", "orange"), color("5xx", "red")] },
  ),
  series(
    "Авторизации: успешные и неуспешные (за минуту)",
    [t("sum by (result) (increase(auth_attempts_total[1m]))", "{{result}}")],
    { x: 8, y: 14, w: 8, h: 9 },
    {
      draw: "bars",
      overrides: [color("success", "green"), color("failure", "red")],
    },
  ),
  series(
    "Отказы в доступе по причине (за минуту)",
    [t("sum by (reason) (increase(auth_denied_total[1m]))", "{{reason}}")],
    { x: 16, y: 14, w: 8, h: 9 },
    {
      draw: "bars",
      overrides: [
        color("no_token", "yellow"),
        color("invalid_token", "purple"),
        color("forbidden", "red"),
      ],
    },
  ),
  {
    id: ++id,
    type: "bargauge",
    title: "Запросы по маршрутам и кодам ответа (всего)",
    datasource: ds,
    gridPos: { x: 0, y: 23, w: 24, h: 13 },
    targets: [
      {
        ...t(
          "sort_desc(sum by (route, status) (http_requests_total))",
          "{{status}} {{route}}",
        ),
        instant: true,
      },
    ],
    fieldConfig: {
      defaults: { unit: "short", color: { mode: "continuous-BlPu" } },
      overrides: [],
    },
    options: {
      orientation: "horizontal",
      displayMode: "gradient",
      showUnfilled: true,
      reduceOptions: { calcs: ["lastNotNull"], fields: "", values: false },
    },
  },
  {
    id: ++id,
    type: "stat",
    title: "Доля попаданий в кэш (5 мин)",
    datasource: ds,
    gridPos: { x: 0, y: 36, w: 6, h: 9 },
    targets: [
      t(
        'sum(rate(cache_operations_total{result="hit"}[5m])) / sum(rate(cache_operations_total{result=~"hit|miss|bypass"}[5m]))',
      ),
    ],
    fieldConfig: {
      defaults: {
        unit: "percentunit",
        decimals: 1,
        min: 0,
        max: 1,
        color: { mode: "thresholds" },
        thresholds: {
          mode: "absolute",
          steps: [
            { color: "red", value: null },
            { color: "orange", value: 0.5 },
            { color: "green", value: 0.8 },
          ],
        },
      },
      overrides: [],
    },
    options: {
      reduceOptions: { calcs: ["lastNotNull"], fields: "", values: false },
      colorMode: "background",
      graphMode: "area",
      textMode: "value",
      justifyMode: "center",
      orientation: "auto",
    },
  },
  series(
    "Операции с кэшем по результату (в секунду)",
    [t("sum by (result) (rate(cache_operations_total[1m]))", "{{result}}")],
    { x: 6, y: 36, w: 18, h: 9 },
    {
      unit: "ops",
      legend: "right",
      overrides: [
        color("hit", "green"),
        color("miss", "orange"),
        color("set", "blue"),
        color("invalidate", "purple"),
        color("bypass", "yellow"),
        color("error", "red"),
      ],
    },
  ),
];
const dashboard = {
  uid: "sportreg-overview",
  title: "SportReg — API",
  tags: ["sportreg"],
  timezone: "browser",
  refresh: "5s",
  schemaVersion: 39,
  version: 1,
  time: { from: "now-15m", to: "now" },
  panels,
};
fs.writeFileSync(process.argv[2], JSON.stringify(dashboard, null, 2) + "\n");
console.log("panels", panels.length);

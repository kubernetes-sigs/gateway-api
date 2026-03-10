function parseFeatureCount(value) {
  var match = value.trim().match(/^(\d+)\/(\d+)$/)
  return match ? parseInt(match[1], 10) : NaN
}

if (!window.gatewayApiFeatureCountSortRegistered) {
  Tablesort.extend(
    "feature-count",
    function(item) {
      return /^\d+\/\d+$/.test(item.trim())
    },
    function(a, b) {
      return parseFeatureCount(a) - parseFeatureCount(b)
    },
  )
  window.gatewayApiFeatureCountSortRegistered = true
}

document$.subscribe(function() {
  var tables = document.querySelectorAll("article table:not([class])")
  tables.forEach(function(table) {
    new Tablesort(table)
  })
})


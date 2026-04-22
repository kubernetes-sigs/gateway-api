function parseFeatureCount(value) {
  var match = value.trim().match(/^(\d+)\/(\d+)$/)
  return match ? parseInt(match[1], 10) : NaN
}

function parseFeatureSummary(value) {
  var text = value.replace(/\s+/g, " ").trim()
  var match = text.match(/(\d+)(?:\/(\d+))?\s+features/i)
  var supported = match ? parseInt(match[1], 10) : Number.MAX_SAFE_INTEGER
  var total = match && match[2] ? parseInt(match[2], 10) : supported
  var status = 0

  if (/Failing/i.test(text)) {
    status = 2
  } else if (/Partially conformant/i.test(text)) {
    status = 1
  }

  return {
    status: status,
    supported: supported,
    total: total,
  }
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

if (!window.gatewayApiFeatureSummarySortRegistered) {
  Tablesort.extend(
    "feature-summary",
    function(item) {
      return /features/i.test(item)
    },
    function(a, b) {
      var aSummary = parseFeatureSummary(a)
      var bSummary = parseFeatureSummary(b)

      if (aSummary.status !== bSummary.status) {
        return bSummary.status - aSummary.status
      }
      if (aSummary.supported !== bSummary.supported) {
        return aSummary.supported - bSummary.supported
      }
      return aSummary.total - bSummary.total
    },
  )
  window.gatewayApiFeatureSummarySortRegistered = true
}

document$.subscribe(function() {
  var tables = document.querySelectorAll("article table:not([class])")
  tables.forEach(function(table) {
    var headers = table.querySelectorAll("thead th")
    headers.forEach(function(header) {
      if (header.textContent.trim() === "Features" || header.textContent.trim() === "Extended Features") {
        header.setAttribute("data-sort-method", "feature-summary")
      }
    })
    new Tablesort(table)
  })
})

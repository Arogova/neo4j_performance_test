inputElement = document.getElementById("input");
inputElement.addEventListener("change", handleFiles, false);
function handleFiles() {
  filePath = this.files[0].name
  Papa.parse(this.files[0], {
    skipEmptyLines: true,
    header: true,
    transformHeader: function (h) {
      return h.trim();
    },
    complete: function (results) {
      // if (document.getElementById("subsetsum-checkbox").checked){
      //   distinctNodes = getDistinctNodes(results)
      //   createSubsetSumChart(format_subsetsum_avg(results), distinctNodes, "subsetsum-plot")
      // } else {
        //nodePos = getNodeToPosMap(results)
        distinctNodes = getDistinctNodes(results)
  
        //timeout_results = compute_timeouts(results)
  
        // createAverageHeatmap(format_average(nodePos, compute_average(results)), distinctNodes, "success-plot", filePath)
        // createMedianHeatmap(format_median(nodePos, compute_median(results)), distinctNodes, "median-plot", filePath)
        // createTimeoutHeatmap(format_timeouts(nodePos, timeout_results), distinctNodes, "timeouts-plot", filePath)
        // createMinNodeLineChart(format_min_nodes(timeout_results), "min-timeout-plot", filePath)

        createMedianAndTimeoutChart(formatMedianLineChart(results), formatTimeoutBarChart(results), distinctNodes, "plot", filePath)
        magic()
      //}
    }
  });
}

function createAverageLineChart(formatted_results, distinctNodes, divId, filePath) {
  options = {
    title: {
      text: "Average execution time of " + getFullProblemName(filePath) + " in ms",
      textStyle: {
        fontSize: 24,
        lineHeight: 10,
      },
      padding: [20, 0, 0, 0],
      left: 'center'
    },
    xAxis: {
      type: "category",
      name: "nodes",
      nameTextStyle: {
        fontSize: 24
      },
      data: distinctNodes,
      axisLabel: {
        fontSize: 24
      },
      offset: 20
    },
    yAxis: {
      type: "value",
      name: "avg exec time in ms",
      nameTextStyle: {
        fontSize: 24
      },
      axisLabel: {
        fontSize: 24
      }
    },
    series: {
      type: "line",
      smooth: true,
      data: formatted_results,
      label: {
        show: true,
        position: 'top',
        fontSize: 20
      }
    },
    toolbox: {
      show: true,
      feature: {
        saveAsImage: {
          show: true,
          name: "avg_exec_line",
          type: "svg"
        }
      }
    }
  }

  chart = echarts.init(document.getElementById(divId), null, { renderer: "svg" })
  chart.setOption(options)
}

function createMedianAndTimeoutChart (median_results, timeout_results, distinct_nodes, divId, filePath) {
  options = {
    // title: {
    //   text: "Median execution time in ms and number of timeouts for " + getFullProblemName(filePath),
    //   textStyle: {
    //     fontSize: 30,
    //     lineHeight: 10,
    //   },
    //   padding: [20, 0, 0, 0],
    //   left: 'center'
    // },
    xAxis: {
      type: "category",
      name: "nodes",
      nameTextStyle: {
        fontSize: 30,
      },
      data: distinct_nodes,
      axisLabel: {
        fontSize: 30,
        offset: 10
      },
      offset: 0
    },
    yAxis: [
      {
      type: "value",
      name: "median execution time in ms",
      min: 0,
      max: 300000,
      nameTextStyle: {
        fontSize: 30
      },
      axisLabel: {
        fontSize: 30
      },
      axisLine: {
        show: true,
        lineStyle: {
          //color: 'rgb(0,114,178)',
          color: 'rgb(0,158,115)',
          width: 5,
          type: 'solid'
        }
      }
    },
    {
      type: "value",
      name: "percentage of timeouts",
      min: 0,
      max: 100,
      interval: 20,
      nameTextStyle: {
        fontSize: 30
      },
      axisLabel: {
        fontSize: 30,
        formatter: function(value, index){
          return parseInt(value) + '%'
        }
      },
      axisLine: {
        show: true,
        lineStyle: {
          //color: 'rgb(213,94,0)',
          color: 'rgb(204,121,167)',
          width: 5,
          type: 'solid'
        }
      }
    }
  ],
    series: [{
      type: "line",
      smooth: true,
      data: median_results,
      yAxisIndex: 0,
      label: {
        show: true,
        position: 'top',
        fontSize: 20
      },
      lineStyle: {
        //color: 'rgb(0,114,178)',
        color: 'rgb(0,158,115)',
        width: 10,
        type: 'solid',
        cap: 'round'
      }
    },
    { 
      type: "bar",
      data: timeout_results,
      yAxisIndex: 1,
      itemStyle: {
        //color: 'rgb(213,94,0)',
        color: 'rgb(204,121,167)',
      }
    }
  ],
    toolbox: {
      show: true,
      feature: {
        saveAsImage: {
          show: true,
          name: "avg_exec_line",
          type: "svg"
        }
      }
    }
  }

  chart = echarts.init(document.getElementById(divId), null, { renderer: "svg" })
  chart.setOption(options)
}

function formatMedianLineChart(results){
  pre_format = []
  element_position = new Map()
  results.data.forEach(res => {
    if (!element_position.has(res["order"])) {
      pre_format.push([])
      element_position.set(res["order"], pre_format.length-1)
    }

    elementPos = element_position.get(res["order"])

    if (res["query execution time"] !== "timeout"  && res["query execution time"] !== "OOM") {
      pre_format[elementPos].push(parseInt(res["query execution time"]))
    }
  })

  console.log(pre_format)

  formatted = []
  pre_format.forEach(el => {
    el.sort((a, b) => a - b)
    half = Math.floor(el.length/2)
    if (el.length == 0) {
      median = -1
    } else if (el.length % 2 == 1) {
      median = el[half]
    } else {
      median = (el[half-1] + el[half])/2
    }
    formatted.push(median)
  })

  console.log(formatted)

  if (formatted.includes(-1) ) return formatted.slice(0, formatted.indexOf(-1))
  else return formatted
}

function formatTimeoutBarChart(results){
  formatted = []
  element_position = new Map()
  results.data.forEach(res => {
    if (!element_position.has(res["order"])) {
      formatted.push(100)
      element_position.set(res["order"], formatted.length-1)
    }

    elementPos = element_position.get(res["order"])

    if (res["query execution time"] !== "timeout"  && res["query execution time"] !== "OOM") {
      formatted[elementPos] = formatted[elementPos]-5
    }
  })

  return formatted
}

function createAverageHeatmap(formatted_average, distinctNodes, divId, filePath) {
  options = {
    title: {
      text: "Average execution time of " + getFullProblemName(filePath) + " in ms",
      textStyle: {
        fontSize: 24,
        lineHeight: 10
      },
      left: 'center',
      padding: [20, 0, 0, 0],
    },
    xAxis: {
      name: "nodes",
      nameTextStyle: {
        fontSize: 24
      },
      type: "category",
      data: distinctNodes,
      axisLabel: {
        fontSize: 24
      },
      splitArea: {
        show: true
      },
      offset: 20,
    },
    yAxis: {
      name: "probability",
      nameTextStyle: {
        fontSize: 24
      },
      type: "category",
      data: [0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9, 1.0],
      axisLabel: {
        fontSize: 24
      },
      splitArea: {
        show: true
      }
    },
    visualMap: {
      min: 0,
      max: 300000,
      calculable: true,
      inRange: {
        color: ["#66bb6a", "#42a5f5"]
      },
      right: "20px",
      top: "10%",
      itemHeight: 550
    },
    series: [{
      name: '',
      type: "heatmap",
      data: formatted_average,
      coordinateSystem: "cartesian2d",
      label: {
        show: true,
        align: 'center',
        fontSize: 20
      }
    }],
    toolbox: {
      show: true,
      feature: {
        saveAsImage: {
          show: true,
          name: "average_execution_time",
          type: "svg"
        }
      }
    }
  }

  averageChart = echarts.init(document.getElementById(divId), null, { renderer: "svg" })
  averageChart.setOption(options)
}

function createMedianHeatmap(formatted_median, distinctNodes, divId, filePath) {
  options = {
    title: {
      text: "Median execution time of " + getFullProblemName(filePath) + " in ms",
      textStyle: {
        fontSize: 24,
        lineHeight: 10
      },
      left: 'center',
      padding: [20, 0, 0, 0],
    },
    xAxis: {
      name: "nodes",
      nameTextStyle: {
        fontSize: 24
      },
      type: "category",
      data: distinctNodes,
      axisLabel: {
        fontSize: 24
      },
      splitArea: {
        show: true
      },
      offset: 20,
    },
    yAxis: {
      name: "probability",
      nameTextStyle: {
        fontSize: 24
      },
      type: "category",
      data: [0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9, 1.0],
      axisLabel: {
        fontSize: 24
      },
      splitArea: {
        show: true
      }
    },
    visualMap: {
      min: 0,
      max: 300000,
      calculable: true,
      inRange: {
        color: ["#66bb6a", "#42a5f5"]
      },
      right: "20px",
      top: "10%",
      itemHeight: 550
    },
    series: [{
      name: '',
      type: "heatmap",
      data: formatted_median,
      coordinateSystem: "cartesian2d",
      label: {
        show: true,
        align: 'center',
        fontSize: 20
      }
    }],
    toolbox: {
      show: true,
      feature: {
        saveAsImage: {
          show: true,
          name: "median_execution_time",
          type: "svg"
        }
      }
    }
  }

  medianChart = echarts.init(document.getElementById(divId), null, { renderer: "svg" })
  medianChart.setOption(options)
}


function createTimeoutHeatmap(formatted_timeouts, distinctNodes, divId, filePath) {
  options = {
    title: {
      text: "Timeout percentage of " + getFullProblemName(filePath),
      textStyle: {
        fontSize: 24,
        lineHeight: 10
      },
      left: 'center',
      padding: [20, 0, 0, 0],
    },
    xAxis: {
      name: "nodes",
      nameTextStyle: {
        fontSize: 24
      },
      type: "category",
      data: distinctNodes,
      axisLabel: {
        fontSize: 24
      },
      splitArea: {
        show: true
      },
      offset: 20
    },
    yAxis: {
      name: "probability",
      nameTextStyle: {
        fontSize: 24
      },
      type: "category",
      data: [0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9, 1.0],
      axisLabel: {
        fontSize: 24
      },
      splitArea: {
        show: true
      }
    },
    visualMap: {
      min: 0,
      max: 100,
      calculable: true,
      inRange: {
        color: ["#66bb6a", "#ec7063"]
      },
      right: "20px",
      top: "10%",
      itemHeight: 550
    },
    series: [{
      name: '',
      type: "heatmap",
      data: formatted_timeouts,
      coordinateSystem: "cartesian2d",
      label: {
        show: true,
        align: 'center',
        fontSize: 20
      }
    }],
    toolbox: {
      show: true,
      feature: {
        saveAsImage: {
          show: true,
          name: "timeout_percentage",
          type: "svg"
        }
      }
    }
  }

  timeoutChart = echarts.init(document.getElementById(divId), null, { renderer: "svg" })
  timeoutChart.setOption(options)
}

function createMinNodeLineChart(formatted_min_node, divId, filePath) {
  options = {
    title: {
      text: "Minimum amount of nodes to 50% timeout for the " + getFullProblemName(filePath) + " problem",
      textStyle: {
        fontSize: 24,
        lineHeight: 10,
      },
      padding: [20, 0, 0, 0],
      left: 'center'
    },
    xAxis: {
      type: "category",
      name: "probability",
      nameTextStyle: {
        fontSize: 24
      },
      data: [0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9, 1.0],
      axisLabel: {
        fontSize: 24
      },
      offset: 20
    },
    yAxis: {
      type: "value",
      name: "minimum nodes to timeout",
      nameTextStyle: {
        fontSize: 24
      },
      axisLabel: {
        fontSize: 24
      }
    },
    series: {
      type: "line",
      smooth: true,
      data: formatted_min_node,
      label: {
        show: true,
        position: 'bottom',
        fontSize: 20
      }
    },
    toolbox: {
      show: true,
      feature: {
        saveAsImage: {
          show: true,
          name: "min_node_to_timeout",
          type: "svg"
        }
      }
    }
  }

  minNodeChart = echarts.init(document.getElementById(divId), null, { renderer: "svg" })
  minNodeChart.setOption(options)
}

function createSubsetSumChart(formatted_results, distinctNodes, divId) {
  options = {
    title: {
      text: "Average execution time for the SubsetSum problem",
      textStyle: {
        fontSize: 24,
        lineHeight: 10,
      },
      padding: [20, 0, 0, 0],
      left: 'center'
    },
    xAxis: {
      type: "category",
      name: "nodes",
      nameTextStyle: {
        fontSize: 24
      },
      data: distinctNodes,
      axisLabel: {
        fontSize: 24
      },
      offset: 20
    },
    yAxis: {
      type: "value",
      name: "avg exec time in ms",
      nameTextStyle: {
        fontSize: 24
      },
      axisLabel: {
        fontSize: 24
      }
    },
    series: {
      type: "line",
      smooth: true,
      data: formatted_results,
      label: {
        show: true,
        position: 'top',
        fontSize: 20
      }
    },
    toolbox: {
      show: true,
      feature: {
        saveAsImage: {
          show: true,
          name: "subsetsum",
          type: "svg"
        }
      }
    }
  }

  subsetSumChart = echarts.init(document.getElementById(divId), null, { renderer: "svg" })
  subsetSumChart.setOption(options)
  console.log(formatted_results)
}

function compute_average(results) {
  unformatted = []
  element_position = new Map()
  results.data.forEach(res => {
    if (!element_position.has(res["order"] + ":" + res["edge probability"])) {
      unformatted.push({
        order: res["order"],
        probability: res["edge probability"],
        avgExecTime: -1
      })
      element_position.set(res["order"] + ":" + res["edge probability"], unformatted.length - 1)
    }

    elementPos = element_position.get(res["order"] + ":" + res["edge probability"])

    if (res["query execution time"] !== "timeout" && res["query execution time"] !== "OOM") {
      if (unformatted[elementPos].avgExecTime == -1) {
        unformatted[elementPos].avgExecTime = res["query execution time"]
      } else {
        unformatted[elementPos].avgExecTime = (parseFloat(unformatted[elementPos].avgExecTime) + parseFloat(res["query execution time"])) / 2
      }
    }
  });
  return unformatted
}

function format_average(nodePos, average_results) {
  formatted = []
  average_results.forEach(el => {
    if (el.avgExecTime == -1) {
      formatted.push({
        value: [nodePos.get(el.order), probPos.get(el.probability), -1],
        itemStyle: { color: "#ec7063" },
        symbol: "circle"
      })
    } else {
      formatted.push({ value: [nodePos.get(el.order), probPos.get(el.probability), Math.round(parseFloat(el.avgExecTime))] })
    }
  })
  return formatted
}

function compute_median(results) {
  unformatted = []
  element_position = new Map()
  results.data.forEach(res => {
    if (!element_position.has(res["order"] + ":" + res["edge probability"])) {
      unformatted.push({
        order: res["order"],
        probability: res["edge probability"],
        medExecTime: []
      })
      element_position.set(res["order"] + ":" + res["edge probability"], unformatted.length - 1)
    }

    elementPos = element_position.get(res["order"] + ":" + res["edge probability"])

    if (res["query execution time"] !== "timeout" && res["query execution time"] !== "OOM") {
      unformatted[elementPos].medExecTime.push(parseInt(res["query execution time"]))
    }
  });

  unformatted.forEach(el => {
    el.medExecTime.sort((a, b) => a - b)
    if (el.medExecTime.length == 0) {
      median = -1
    } else if (el.medExecTime.length % 2 == 1) {
      median = el.medExecTime[(el.medExecTime.length+1)/2]
    } else {
      median = (el.medExecTime[el.medExecTime.length/2] + el.medExecTime[(el.medExecTime.length/2)+1])/2
    }
    el.medExecTime = median
  })
  return unformatted
}

function format_median(nodePos, average_results) {
  formatted = []
  average_results.forEach(el => {
    if (el.medExecTime == -1) {
      formatted.push({
        value: [nodePos.get(el.order), probPos.get(el.probability), -1],
        itemStyle: { color: "#ec7063" },
        symbol: "circle"
      })
    } else {
      formatted.push({ value: [nodePos.get(el.order), probPos.get(el.probability), Math.round(parseFloat(el.medExecTime))] })
    }
  })
  return formatted
}

function compute_timeouts(results) {
  unformatted = []
  element_position = new Map()
  results.data.forEach(res => {
    if (!element_position.has(res["order"] + ":" + res["edge probability"])) {
      unformatted.push({
        order: res["order"],
        probability: res["edge probability"],
        percTimeout: 100
      })
      element_position.set(res["order"] + ":" + res["edge probability"], unformatted.length - 1)
    }

    elementPos = element_position.get(res["order"] + ":" + res["edge probability"])


    if (res["query execution time"] !== "timeout" && res["query execution time"] !== "OOM") {
      unformatted[elementPos].percTimeout = unformatted[elementPos].percTimeout - 5
    }
  });

  return unformatted
}

function format_timeouts(nodePos, timeout_results) {
  formatted = timeout_results.map(el => {
    return [nodePos.get(el.order), probPos.get(el.probability), el.percTimeout]
  })
  return formatted
}

function format_min_nodes(timeoutResults) {
  formatted = []
  element_position = new Map()

  timeoutResults.forEach(res => {
    if (!element_position.has(res.probability)) {
      formatted.push("-")
      element_position.set(res.probability, formatted.length - 1)
    }

    elementPos = element_position.get(res.probability)

    if (res.percTimeout >= 50 && (formatted[elementPos] == "-" || parseInt(formatted[elementPos]) > parseInt(res.order))) {
      formatted[elementPos] = res.order
    }
  })

  return formatted
}

function format_subsetsum_avg (results) {
  formatted = []
  element_position = new Map()
  results.data.forEach(res => {
    if (!element_position.has(res["order"])) {
      formatted.push("-")
      element_position.set(res["order"], formatted.length-1)
    }

    elementPos = element_position.get(res["order"])

    if (res["query execution time"] !== "timeout"  && res["query execution time"] !== "OOM") {
      if (formatted[elementPos] == "-") {
        formatted[elementPos] = res["query execution time"]
      } else {
        formatted[elementPos] - (parseFloat(formatted[elementPos])) + parseFloat(res["query execution time"]) / 2
      }
    }
  })

  return formatted
}

function getNodeToPosMap(results) {
  distinctNodes = getDistinctNodes(results)
  distinctNodes.sort((a, b) => parseInt(a) - parseInt(b))
  nodePos = new Map()
  lastNodePos = 0
  distinctNodes.forEach(el => {
    nodePos.set(el, lastNodePos)
    lastNodePos++
  })

  return nodePos
}

function getDistinctNodes(results) {
  distinctNodes = []
  results.data.forEach(res => {
    if (!distinctNodes.includes(res["order"])) distinctNodes.push(res["order"])
  })

  return distinctNodes
}

function getFullProblemName(filePath) {
  problems = new Map([
    ["hamil", "Hamiltonian path"],
    ["euler", "Eulerian path"],
    ["subset", "Subset sum"],
    ["Star", "A*BA*"]
  ])

  problemName = "unknown problem"
  problems.forEach((fullName, shortName, _) => {
    if (filePath.includes(shortName)) {
      problemName = fullName
    }
  })

  return problemName
}

probPos = new Map([['0.1', 0], ['0.2', 1], ['0.3', 2], ['0.4', 3], ['0.5', 4], ['0.6', 5], ['0.7', 6], ['0.8', 7], ['0.9', 8], ['1.0', 9]])

function magic() {
  nullValues = Array.from(document.getElementsByTagName("text"))
  nullValues.forEach((el) => {
    if (el.innerHTML === "-1") {
      el.innerHTML = "&infin;"
    }
  })
}
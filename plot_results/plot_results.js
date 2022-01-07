function analyze_data (results) {
  success_trace = {
    x: [],
    y: [],
    type : 'scatter'
  }
  timeout_trace = {
    x: [],
    y: [],
    type : 'scatter'
  }
  results.data.forEach(row => {
    order = parseFloat(row['order']);
    i = (order/10)-1;
    time = row['query execution time'].trim();
    if (success_trace.x[i] === undefined) {
      success_trace.x[i] = order;
      timeout_trace.x[i] = order;
    }
    if (timeout_trace.y[i] === undefined) timeout_trace.y[i] = 0;
    if (success_trace.y[i] === undefined) success_trace.y[i] = 0;
    if (time.localeCompare('timeout') === 0) timeout_trace.y[i] += 1;
    else success_trace.y[i] = (success_trace.y[i] + parseFloat(time))/2;
  });
  return {
    success_trace: success_trace,
    timeout_trace: timeout_trace
  }
}

function plot_success (success_data){
  console.log(success_data);
  success_plot_layout = {
    title: 'Average run time of successful runs',
    showlegend: false,
    autosize: true,
    width: 800,
    height: 800,
    xaxis : {title: 'Number of nodes'},
    yaxis : {title: 'Average run time in ms'}
  };


  Plotly.newPlot('success-plot', [success_data], success_plot_layout);
}

function plot_timeouts (timeouts_data){
  timeouts_plot_layout = {
    title: '# of timeouts per scenario (out of 10 runs) ',
    showlegend: false,
    autosize: true,
    width: 800,
    height: 800,
    xaxis : {title: 'Numer of nodes'},
    yaxis : {title: '# of timeouts'}
  };


  Plotly.newPlot('timeouts-plot', [timeouts_data], timeouts_plot_layout);
}

inputElement = document.getElementById("input");
inputElement.addEventListener("change", handleFiles, false);
function handleFiles() {
  Papa.parse(this.files[0], {
    header: true,
    transformHeader:function(h) {
      return h.trim();
    },
    complete: function(results) {
      analyzed = analyze_data(results);
      console.log(analyzed);
      plot_success(analyzed.success_trace);
      plot_timeouts(analyzed.timeout_trace);
    }
  });
}

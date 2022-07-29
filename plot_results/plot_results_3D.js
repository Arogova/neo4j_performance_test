save_as_svg_config = {
  modeBarButtonsToRemove: ['toImage', 'sendDataToCloud'],
  modeBarButtonsToAdd: [{
    name: 'toImage2',
    icon: Plotly.Icons.camera,
    click: function(gd) {
      Plotly.downloadImage(gd, {format: 'svg'})
    }
  }]
}

function analyze_data (results){
  success_traces = [];
  timeout_traces = [];

  current_probability = 0.1;
  current_order = 10;
  current_avg_time = 0.0;
  current_timeout_counter = 0;

  results.data.forEach(row => {
    order = parseFloat(row['order']);
    probability = parseFloat(row['edge probability']);
    time = row['query execution time'];

    if (order !== current_order) {
      i = (current_probability*10)-1;
      if (success_traces[i] === undefined) {
        new_success_trace = {
            x: [],
            y: [],
            z: [],
            mode: 'lines',
            line : {
              color: '#1f77b4',
              width: 1
            },
            type: 'scatter3d'
        };
        new_timeout_trace = {
            x: [],
            y: [],
            z: [],
            mode: 'lines',
            line : {
              color: '#1f77b4',
              width: 1
            },
            type: 'scatter3d'
        };
        success_traces.push(new_success_trace);
        timeout_traces.push(new_timeout_trace);
      }
      success_traces[i].x.push(current_probability);
      success_traces[i].y.push(current_order);
      success_traces[i].z.push(current_avg_time);
      timeout_traces[i].x.push(current_probability);
      timeout_traces[i].y.push(current_order);
      timeout_traces[i].z.push(current_timeout_counter);
      current_probability = probability;
      current_order = order;
      current_avg_time = 0.0;
      current_timeout_counter = 0;
    }

    if (time.localeCompare("timeout") !== 0){
      if (current_avg_time == 0.0) current_avg_time = parseFloat(time);
      else current_avg_time = (current_avg_time + parseFloat(time))/2;
    } else {
      current_timeout_counter++;
    }
  });

  return {
    success: success_traces,
    timeouts: timeout_traces
  }
}

function plot_success (success_data){
  success_plot = document.getElementById('success-plot');
  success_plot_layout = {
    title: 'Average run time of successful runs',
    showlegend: false,
    autosize: true,
    width: 800,
    height: 800,
    scene: {
      xaxis : {title: 'Edge probability'},
      yaxis : {title: 'Numer of nodes'},
      zaxis : {title: 'Average run time in ms'}
    }
  };

  Plotly.newPlot(success_plot, success_data, success_plot_layout, save_as_svg_config);
}

function plot_timeouts (timeouts_data){
  timeouts_plot = document.getElementById('timeouts-plot');
  timeouts_plot_layout = {
    title: '# of timeouts per scenario (out of 10 runs) ',
    showlegend: false,
    autosize: true,
    width: 800,
    height: 800,
    scene: {
      xaxis : {title: 'Edge probability'},
      yaxis : {title: 'Numer of nodes'},
      zaxis : {title: '# of timeouts'}
    }
  };

    Plotly.newPlot(timeouts_plot, timeouts_data, timeouts_plot_layout, save_as_svg_config);
}

inputElement = document.getElementById("input");
inputElement.addEventListener("change", handleFiles, false);
function handleFiles() {
  Papa.parse(this.files[0], {
    skipEmptyLines: true,
    header: true,
    transformHeader:function(h) {
      return h.trim();
    },
    complete: function(results) {
      analyzed = analyze_data(results);
      plot_success(analyzed.success);
      plot_timeouts(analyzed.timeouts);
    }
  });
}

// Color palette for charts
export const CHART_COLORS = {
    primary: '#8884d8',
    secondary: '#82ca9d',
    tertiary: '#ffc658',
    quaternary: '#ff7300',
    // Additional colors for various chart elements
    series: [
      '#8884d8',
      '#82ca9d',
      '#ffc658',
      '#ff7300',
      '#0088fe',
      '#00c49f'
    ]
  };
  
  // Chart dimensions and responsive breakpoints
  export const CHART_DIMENSIONS = {
    default: {
      height: 300,
      width: '100%'
    },
    compact: {
      height: 200,
      width: '100%'
    }
  };
  
  // Common margin configuration for charts
  export const CHART_MARGINS = {
    default: {
      top: 5,
      right: 30,
      left: 20,
      bottom: 5
    },
    withLegend: {
      top: 5,
      right: 30,
      left: 20,
      bottom: 25
    }
  };
  
  // Configuration for different chart types
  export const CHART_CONFIG = {
    line: {
      strokeWidth: 2,
      dot: {
        radius: 4,
        strokeWidth: 2
      },
      activeDot: {
        radius: 6,
        strokeWidth: 2
      }
    },
    bar: {
      barSize: 20,
      radius: [4, 4, 0, 0]
    },
    pie: {
      innerRadius: 60,
      outerRadius: 80,
      paddingAngle: 2,
      cornerRadius: 4
    }
  };
  
  // Tooltip styling configuration
  export const TOOLTIP_STYLES = {
    contentStyle: {
      backgroundColor: 'rgba(255, 255, 255, 0.9)',
      border: '1px solid #ccc',
      borderRadius: '4px',
      padding: '8px'
    },
    labelStyle: {
      color: '#666',
      fontWeight: 500
    },
    itemStyle: {
      color: '#333',
      padding: '2px 0'
    }
  };
  
  // Common axis configuration
  export const AXIS_CONFIG = {
    xAxis: {
      tickMargin: 10,
      tickSize: 5,
      tickLine: false,
      axisLine: {
        stroke: '#E5E7EB'
      },
      tick: {
        fontSize: 12,
        fill: '#6B7280'
      }
    },
    yAxis: {
      tickMargin: 10,
      tickSize: 5,
      tickLine: false,
      axisLine: {
        stroke: '#E5E7EB'
      },
      tick: {
        fontSize: 12,
        fill: '#6B7280'
      },
      // Allow y-axis to start from zero for better data visualization
      allowZero: true
    }
  };
  
  // Animation configuration
  export const ANIMATION_CONFIG = {
    initial: {
      duration: 800,
      easing: 'ease-out'
    },
    update: {
      duration: 500,
      easing: 'ease-in-out'
    }
  };
  
  // Legend configuration
  export const LEGEND_CONFIG = {
    align: 'center',
    verticalAlign: 'bottom',
    layout: 'horizontal',
    iconSize: 10,
    iconType: 'circle',
    wrapperStyle: {
      paddingTop: '10px'
    }
  };
  
  // Grid line configuration
  export const GRID_CONFIG = {
    horizontal: {
      strokeDasharray: '3 3',
      stroke: '#E5E7EB',
      opacity: 0.5
    },
    vertical: {
      strokeDasharray: '3 3',
      stroke: '#E5E7EB',
      opacity: 0.5
    }
  };
  
  // Time formats for different chart scales
  export const TIME_FORMATS = {
    hourly: 'HH:mm',
    daily: 'MMM dd',
    monthly: 'MMM yyyy',
    yearly: 'yyyy'
  };
  
  // Export a preconfigured chart theme that combines multiple settings
  export const CHART_THEME = {
    colors: CHART_COLORS,
    dimensions: CHART_DIMENSIONS,
    margins: CHART_MARGINS,
    tooltip: TOOLTIP_STYLES,
    axis: AXIS_CONFIG,
    animation: ANIMATION_CONFIG,
    legend: LEGEND_CONFIG,
    grid: GRID_CONFIG
  };
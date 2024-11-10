const childProcess = require('child_process');

const apps = {
  'session-management': {
    start: 'cd ../session-management/cmd/server && go run main.go --port 9001 &',
    stop: 'lsof -ti:9001 | xargs kill -9'
  },
  leaderboard: {
    start: 'cd ../leaderboard/cmd/server && go run main.go --port 9002 &',
    stop: 'lsof -ti:9002 | xargs kill -9'
  },
  'sales-analytics': {
    start: 'cd ../sales-analytics/cmd/server && go run main.go --port 9003 &',
    stop: 'lsof -ti:9003 | xargs kill -9'
  },
  luckydraw: {
    start: 'cd ../lucky-draw/cmd/server && go run main.go --port 9004 &',
    stop: 'lsof -ti:9004 | xargs kill -9'
  },
  frontend: {
    start: 'cd ../../frontend && npm run dev -- --port 5000 &',
    stop: 'lsof -ti:5000 | xargs kill -9'
  }
};

const startApp = (appName) => {
  if (apps[appName]) {
    childProcess.exec(apps[appName].start, (error, stdout, stderr) => {
      if (error) {
        console.error(`Error starting ${appName}: ${error}`);
      } else {
        console.log(`Started ${appName} on port ${getPort(appName)}`);
      }
    });
  } else {
    console.error(`Unknown app: ${appName}`);
  }
};

const stopApp = (appName) => {
  if (apps[appName]) {
    childProcess.exec(apps[appName].stop, (error, stdout, stderr) => {
      if (error) {
        console.error(`Error stopping ${appName}: ${error}`);
      } else {
        console.log(`Stopped ${appName}`);
      }
    });
  } else {
    console.error(`Unknown app: ${appName}`);
  }
};

const getPort = (appName) => {
  const ports = {
    'session-management': 9001,
    leaderboard: 9002,
    'sales-analytics': 9003,
    luckydraw: 9004,
    frontend: 3000
  };
  return ports[appName] || 'unknown port';
};

const startBothApps = () => {
  startApp('frontend');
  startApp('session-management');
};

const args = process.argv.slice(2);

if (args.length < 2) {
  console.error('Usage: node runapp.js {start|stop} {app_name}');
  process.exit(1);
}

const action = args[0];
const appName = args[1];

if (action === 'start') {
  if (appName === 'both') {
    startBothApps();
  } else {
    startApp(appName);
  }
} else if (action === 'stop') {
  stopApp(appName);
} else {
  console.error('Invalid action. Please use "start" or "stop".');
  process.exit(1);
}


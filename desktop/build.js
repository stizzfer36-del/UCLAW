const fs = require('fs');
const path = require('path');
const os = require('os');

const root = __dirname;
const src = path.join(root, 'src');
const out = path.join(root, 'dist');
const electronDir = path.join(root, 'electron');

function writeFile(target, body) {
  fs.mkdirSync(path.dirname(target), { recursive: true });
  fs.writeFileSync(target, body);
}

function copyOrDefault(source, target, fallback) {
  if (fs.existsSync(source)) {
    writeFile(target, fs.readFileSync(source));
    return;
  }
  writeFile(target, fallback);
}

function doctor() {
  const checks = {
    node: process.version,
    platform: os.platform(),
    electron: fs.existsSync(path.join(root, 'node_modules', 'electron')) ? 'installed' : 'missing',
    electron_builder: fs.existsSync(path.join(root, 'node_modules', 'electron-builder')) ? 'installed' : 'missing'
  };
  console.log(JSON.stringify(checks, null, 2));
}

if (process.argv.includes('--doctor')) {
  doctor();
  process.exit(0);
}

fs.mkdirSync(out, { recursive: true });

copyOrDefault(
  path.join(src, 'index.html'),
  path.join(out, 'index.html'),
  '<!doctype html><html><body><h1>UCLAW Desktop</h1><p>Missing desktop/src/index.html</p></body></html>'
);
copyOrDefault(path.join(src, 'app.js'), path.join(out, 'app.js'), 'console.log("UCLAW desktop bootstrap missing");\n');
copyOrDefault(path.join(src, 'styles.css'), path.join(out, 'styles.css'), 'body{font-family:sans-serif}\n');
copyOrDefault(
  path.join(src, 'state.json'),
  path.join(out, 'state.json'),
  JSON.stringify({
    missions: [],
    artifacts: [],
    agents: [],
    approvals: [],
    review_queue: [],
    health: {},
    budget: { tokens: 0 },
    errors: [],
    timeline: [],
    workflow_queue: []
  }, null, 2)
);

writeFile(path.join(electronDir, 'main.js'), `const { app, BrowserWindow } = require('electron');
const path = require('path');

function createWindow() {
  const win = new BrowserWindow({
    width: 1440,
    height: 960,
    backgroundColor: '#f7f1e4',
    webPreferences: {
      preload: path.join(__dirname, 'preload.js')
    }
  });
  win.loadFile(path.join(__dirname, '..', 'dist', 'index.html'));
}

app.whenReady().then(() => {
  createWindow();
  app.on('activate', () => {
    if (BrowserWindow.getAllWindows().length === 0) {
      createWindow();
    }
  });
});

app.on('window-all-closed', () => {
  if (process.platform !== 'darwin') {
    app.quit();
  }
});
`);

writeFile(path.join(electronDir, 'preload.js'), `const { contextBridge } = require('electron');

contextBridge.exposeInMainWorld('uclawDesktop', {
  runtime: 'electron'
});
`);

writeFile(path.join(out, 'README.txt'), 'Desktop assets are rendered by the Go CLI or packaged by the Electron shell.\n');
console.log(out);

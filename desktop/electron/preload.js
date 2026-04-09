const { contextBridge } = require('electron');

contextBridge.exposeInMainWorld('uclawDesktop', {
  runtime: 'electron'
});

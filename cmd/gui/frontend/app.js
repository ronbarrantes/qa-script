// State
let locationsFile = '';
let prioritiesFile = '';
let lastOutputPath = '';

// Initialize
document.addEventListener('DOMContentLoaded', () => {
    // Make drop zones clickable
    document.getElementById('locations-zone').addEventListener('click', () => {
        selectFile('locations');
    });
    document.getElementById('priorities-zone').addEventListener('click', () => {
        selectFile('priorities');
    });

    // Handle Wails native file drop events (required for Windows)
    // The event provides x, y coordinates and files array
    window.runtime?.EventsOn('wails:file-drop', (x, y, files) => {
        if (!files || files.length === 0) return;
        
        const file = files[0];
        const locationsZone = document.getElementById('locations-zone');
        const prioritiesZone = document.getElementById('priorities-zone');
        
        // Get element at drop coordinates to determine which zone
        const dropTarget = document.elementFromPoint(x, y);
        
        // Check if drop was on locations zone
        if (locationsZone.contains(dropTarget) || dropTarget === locationsZone) {
            if (file.toLowerCase().endsWith('.csv')) {
                setFile('locations', file);
            } else {
                showStatus('Please drop a CSV file for locations', 'error');
                showZoneError('locations');
            }
            return;
        }
        
        // Check if drop was on priorities zone
        if (prioritiesZone.contains(dropTarget) || dropTarget === prioritiesZone) {
            if (file.toLowerCase().endsWith('.xlsx')) {
                setFile('priorities', file);
            } else {
                showStatus('Please drop an XLSX file for priorities', 'error');
                showZoneError('priorities');
            }
            return;
        }
        
        // File dropped outside specific zones - auto-detect by extension
        if (file.toLowerCase().endsWith('.csv')) {
            setFile('locations', file);
        } else if (file.toLowerCase().endsWith('.xlsx')) {
            setFile('priorities', file);
        } else {
            showStatus('Unsupported file type. Please use CSV or XLSX files.', 'error');
        }
    });
});

// Drag and drop handlers
function handleDragOver(event) {
    event.preventDefault();
    event.currentTarget.classList.add('drag-over');
}

function handleDragLeave(event) {
    event.currentTarget.classList.remove('drag-over');
}

function handleDrop(event, type) {
    event.preventDefault();
    event.currentTarget.classList.remove('drag-over');
    // On Windows, native file drop is handled by Wails via wails:file-drop event
    // The webview's dataTransfer doesn't contain proper file paths on Windows
}

// File selection via dialog
async function selectFile(type) {
    try {
        let path;
        if (type === 'locations') {
            path = await window.go.main.App.SelectLocationsFile();
        } else {
            path = await window.go.main.App.SelectPrioritiesFile();
        }
        
        if (path) {
            updateFileDisplay(type, path);
        }
    } catch (err) {
        showStatus('Error: ' + err, 'error');
        showZoneError(type);
    }
}

// Set file from drag and drop
async function setFile(type, path) {
    try {
        if (type === 'locations') {
            await window.go.main.App.SetLocationsFile(path);
        } else {
            await window.go.main.App.SetPrioritiesFile(path);
        }
        updateFileDisplay(type, path);
    } catch (err) {
        showStatus('Error: ' + err, 'error');
        showZoneError(type);
    }
}

// Update UI to show selected file
function updateFileDisplay(type, path) {
    const zone = document.getElementById(type + '-zone');
    const fileNameEl = document.getElementById(type + '-file');
    
    // Extract just the filename
    const fileName = path.split('/').pop().split('\\').pop();
    fileNameEl.textContent = fileName;
    zone.classList.add('has-file');
    zone.classList.remove('error');
    
    if (type === 'locations') {
        locationsFile = path;
    } else {
        prioritiesFile = path;
    }
    
    updateOkButton();
    clearStatus();
}

// Show error on drop zone
function showZoneError(type) {
    const zone = document.getElementById(type + '-zone');
    zone.classList.add('error');
    setTimeout(() => zone.classList.remove('error'), 400);
}

// Update OK button state
function updateOkButton() {
    const okBtn = document.getElementById('ok-btn');
    okBtn.disabled = !(locationsFile && prioritiesFile);
}

// Reset files
async function resetFiles() {
    await window.go.main.App.Reset();
    
    locationsFile = '';
    prioritiesFile = '';
    
    document.getElementById('locations-file').textContent = '';
    document.getElementById('priorities-file').textContent = '';
    document.getElementById('locations-zone').classList.remove('has-file', 'error');
    document.getElementById('priorities-zone').classList.remove('has-file', 'error');
    
    updateOkButton();
    clearStatus();
}

// Process files
async function processFiles() {
    const okBtn = document.getElementById('ok-btn');
    okBtn.disabled = true;
    showStatus('Processing...', 'processing');
    
    try {
        const outputPath = await window.go.main.App.Process();
        lastOutputPath = outputPath;
        // Extract just the filename for display
        const fileName = outputPath.split('/').pop().split('\\').pop();
        showStatusWithOpen('✓ ' + fileName, outputPath);
    } catch (err) {
        showStatus('Error: ' + err, 'error');
    } finally {
        okBtn.disabled = false;
    }
}

// Open the output file
async function openOutputFile() {
    if (!lastOutputPath) return;
    try {
        await window.go.main.App.OpenFile(lastOutputPath);
    } catch (err) {
        showStatus('Error opening file: ' + err, 'error');
    }
}

// Status display helpers
function showStatus(message, type) {
    const status = document.getElementById('status');
    const statusMessage = document.getElementById('status-message');
    const openBtn = document.getElementById('open-file-btn');
    
    statusMessage.textContent = message;
    status.className = 'status ' + type;
    openBtn.style.display = 'none';
}

function showStatusWithOpen(message, fullPath) {
    const status = document.getElementById('status');
    const statusMessage = document.getElementById('status-message');
    const openBtn = document.getElementById('open-file-btn');
    
    statusMessage.textContent = message;
    statusMessage.title = fullPath; // Show full path on hover
    status.className = 'status success';
    openBtn.style.display = 'inline-block';
}

function clearStatus() {
    const status = document.getElementById('status');
    const statusMessage = document.getElementById('status-message');
    const openBtn = document.getElementById('open-file-btn');
    
    statusMessage.textContent = '';
    statusMessage.title = '';
    status.className = 'status';
    openBtn.style.display = 'none';
    lastOutputPath = '';
}

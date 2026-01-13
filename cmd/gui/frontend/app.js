// State
let locationsFile = '';
let prioritiesFile = '';

// Initialize
document.addEventListener('DOMContentLoaded', () => {
    // Make drop zones clickable
    document.getElementById('locations-zone').addEventListener('click', () => {
        selectFile('locations');
    });
    document.getElementById('priorities-zone').addEventListener('click', () => {
        selectFile('priorities');
    });

    // Handle Wails file drop events
    window.runtime?.EventsOn('wails:file-drop', (files) => {
        if (files && files.length > 0) {
            const file = files[0];
            if (file.endsWith('.csv')) {
                setFile('locations', file);
            } else if (file.endsWith('.xlsx')) {
                setFile('priorities', file);
            }
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
    
    // Get dropped files
    const files = event.dataTransfer.files;
    if (files.length > 0) {
        const file = files[0];
        // For web file drops, we need to use the Wails backend
        // The actual path will be handled by Wails file drop event
    }
    
    // Also check for file path in text (from Wails drop)
    const text = event.dataTransfer.getData('text');
    if (text) {
        setFile(type, text);
    }
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
        showStatus('✓ Output saved to: ' + outputPath, 'success');
    } catch (err) {
        showStatus('Error: ' + err, 'error');
    } finally {
        okBtn.disabled = false;
    }
}

// Status display helpers
function showStatus(message, type) {
    const status = document.getElementById('status');
    status.textContent = message;
    status.className = 'status ' + type;
}

function clearStatus() {
    const status = document.getElementById('status');
    status.textContent = '';
    status.className = 'status';
}

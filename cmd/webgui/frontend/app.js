// State
let locationsFile = '';
let prioritiesFile = '';
let rulesFile = '';

// Initialize
document.addEventListener('DOMContentLoaded', () => {
    const locationsZone = document.getElementById('locations-zone');
    const prioritiesZone = document.getElementById('priorities-zone');
    const rulesZone = document.getElementById('rules-zone');
    const locationsInput = document.getElementById('locations-input');
    const prioritiesInput = document.getElementById('priorities-input');
    const rulesInput = document.getElementById('rules-input');

    // Click to browse - Locations
    locationsZone.addEventListener('click', () => {
        locationsInput.click();
    });

    // Click to browse - Priorities
    prioritiesZone.addEventListener('click', () => {
        prioritiesInput.click();
    });

    // Click to browse - Rules
    rulesZone.addEventListener('click', () => {
        rulesInput.click();
    });

    // File input change handlers
    locationsInput.addEventListener('change', (e) => {
        if (e.target.files.length > 0) {
            uploadFile(e.target.files[0], 'locations');
        }
    });

    prioritiesInput.addEventListener('change', (e) => {
        if (e.target.files.length > 0) {
            uploadFile(e.target.files[0], 'priorities');
        }
    });

    rulesInput.addEventListener('change', (e) => {
        if (e.target.files.length > 0) {
            uploadFile(e.target.files[0], 'rules');
        }
    });

    // Drag and drop - Locations
    setupDropZone(locationsZone, 'locations', ['.csv']);
    
    // Drag and drop - Priorities
    setupDropZone(prioritiesZone, 'priorities', ['.xlsx']);

    // Drag and drop - Rules (accepts .yaml and .yml)
    setupDropZone(rulesZone, 'rules', ['.yaml', '.yml']);

    // Prevent default drag behavior on document
    document.addEventListener('dragover', (e) => e.preventDefault());
    document.addEventListener('drop', (e) => e.preventDefault());
});

// Set up drag and drop for a zone
function setupDropZone(zone, type, extensions) {
    zone.addEventListener('dragover', (e) => {
        e.preventDefault();
        e.stopPropagation();
        zone.classList.add('drag-over');
    });

    zone.addEventListener('dragleave', (e) => {
        e.preventDefault();
        e.stopPropagation();
        zone.classList.remove('drag-over');
    });

    zone.addEventListener('drop', (e) => {
        e.preventDefault();
        e.stopPropagation();
        zone.classList.remove('drag-over');

        const files = e.dataTransfer.files;
        if (files.length > 0) {
            const file = files[0];
            
            // Check file extension (supports multiple extensions)
            const fileName = file.name.toLowerCase();
            const validExtension = extensions.some(ext => fileName.endsWith(ext));
            if (!validExtension) {
                const extList = extensions.map(e => e.toUpperCase().slice(1)).join('/');
                showStatus(`Please drop a ${extList} file for ${type}`, 'error');
                showZoneError(type);
                return;
            }

            uploadFile(file, type);
        }
    });
}

// Upload file to server
async function uploadFile(file, type) {
    const zone = document.getElementById(type + '-zone');
    zone.classList.add('uploading');
    
    try {
        const formData = new FormData();
        formData.append('file', file);

        const endpoints = {
            'locations': '/api/upload-locations',
            'priorities': '/api/upload-priorities',
            'rules': '/api/upload-rules'
        };
        const endpoint = endpoints[type];
        
        const response = await fetch(endpoint, {
            method: 'POST',
            body: formData
        });

        if (!response.ok) {
            const error = await response.text();
            throw new Error(error);
        }

        const result = await response.json();
        updateFileDisplay(type, result.filename);
    } catch (err) {
        showStatus('Error: ' + err.message, 'error');
        showZoneError(type);
    } finally {
        zone.classList.remove('uploading');
    }
}

// Update UI to show selected file
function updateFileDisplay(type, filename) {
    const zone = document.getElementById(type + '-zone');
    const fileNameEl = document.getElementById(type + '-file');
    
    fileNameEl.textContent = filename;
    zone.classList.add('has-file');
    zone.classList.remove('error');
    
    if (type === 'locations') {
        locationsFile = filename;
    } else if (type === 'priorities') {
        prioritiesFile = filename;
    } else if (type === 'rules') {
        rulesFile = filename;
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
    try {
        await fetch('/api/reset', { method: 'POST' });
        
        locationsFile = '';
        prioritiesFile = '';
        rulesFile = '';
        
        document.getElementById('locations-file').textContent = '';
        document.getElementById('priorities-file').textContent = '';
        document.getElementById('rules-file').textContent = '';
        document.getElementById('locations-zone').classList.remove('has-file', 'error');
        document.getElementById('priorities-zone').classList.remove('has-file', 'error');
        document.getElementById('rules-zone').classList.remove('has-file', 'error');
        
        // Reset file inputs
        document.getElementById('locations-input').value = '';
        document.getElementById('priorities-input').value = '';
        document.getElementById('rules-input').value = '';
        
        updateOkButton();
        clearStatus();
    } catch (err) {
        showStatus('Error: ' + err.message, 'error');
    }
}

// Process files
async function processFiles() {
    const okBtn = document.getElementById('ok-btn');
    okBtn.disabled = true;
    showStatus('Processing...', 'processing');
    
    try {
        const response = await fetch('/api/process', { method: 'POST' });
        
        if (!response.ok) {
            const error = await response.text();
            throw new Error(error);
        }

        const result = await response.json();
        
        // Show success and provide download link
        showSuccessWithDownload(result.downloadUrl, result.filename);
    } catch (err) {
        showStatus('Error: ' + err.message, 'error');
        okBtn.disabled = false;
    }
}

// Show success message with download link
function showSuccessWithDownload(url, filename) {
    const status = document.getElementById('status');
    status.className = 'status success';
    status.innerHTML = `
        <span>✓ Processing complete! </span>
        <a href="${url}" download="${filename}" class="download-link">Download ${filename}</a>
    `;
    
    // Auto-trigger download
    const link = document.createElement('a');
    link.href = url;
    link.download = filename;
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    
    // Re-enable button
    document.getElementById('ok-btn').disabled = false;
}

// Shutdown the server
async function shutdown() {
    try {
        await fetch('/api/shutdown', { method: 'POST' });
        showStatus('Application shutting down...', 'processing');
        setTimeout(() => {
            document.body.innerHTML = '<div style="display:flex;justify-content:center;align-items:center;height:100vh;color:#fff;font-family:sans-serif;text-align:center;"><div><p style="font-size:24px;margin-bottom:16px;">✓ Application Closed</p><p style="opacity:0.6;">You can close this tab now.</p></div></div>';
        }, 500);
    } catch (err) {
        // Server might have already shut down
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

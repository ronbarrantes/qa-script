// State management
const state = {
    csv: {
        valid: false,
        path: '',
        fileName: ''
    },
    excel: {
        valid: false,
        path: '',
        fileName: ''
    }
};

// DOM elements - initialized after DOM ready
let csvZone, excelZone, okBtn;

// Initialize drag and drop
function initDragDrop() {
    csvZone = document.getElementById('csv-zone');
    excelZone = document.getElementById('excel-zone');
    okBtn = document.getElementById('ok-btn');
    
    [csvZone, excelZone].forEach(zone => {
        const type = zone.dataset.type;
        
        zone.addEventListener('dragenter', (e) => {
            e.preventDefault();
            e.stopPropagation();
            zone.classList.add('drag-over');
        });
        
        zone.addEventListener('dragover', (e) => {
            e.preventDefault();
            e.stopPropagation();
            zone.classList.add('drag-over');
        });
        
        zone.addEventListener('dragleave', (e) => {
            e.preventDefault();
            e.stopPropagation();
            // Only remove if we're leaving the zone (not entering a child)
            if (!zone.contains(e.relatedTarget)) {
                zone.classList.remove('drag-over');
            }
        });
        
        zone.addEventListener('drop', async (e) => {
            e.preventDefault();
            e.stopPropagation();
            zone.classList.remove('drag-over');
            
            // Get the dropped files
            const files = e.dataTransfer.files;
            if (files.length === 0) {
                // Try to get file path from URI list (for native file drops)
                const uri = e.dataTransfer.getData('text/uri-list');
                if (uri) {
                    let filePath = uri;
                    // Handle file:// protocol
                    if (filePath.startsWith('file://')) {
                        filePath = decodeURIComponent(filePath.substring(7));
                        // On Windows, remove leading slash for paths like /C:/...
                        if (/^\/[A-Za-z]:/.test(filePath)) {
                            filePath = filePath.substring(1);
                        }
                    }
                    await handleFileDrop(type, filePath);
                }
                return;
            }
            
            // In Wails/Electron, we can get the file path from the File object
            const file = files[0];
            
            // Get the file path - Wails provides it via the path property
            let filePath = file.path || file.name;
            
            // On Windows, paths might have backslashes
            if (filePath) {
                await handleFileDrop(type, filePath);
            } else {
                showError(zone, type, 'Could not read file path. Please use the Browse button.');
            }
        });
        
        // Click to open file dialog (on the zone itself, not the browse button which has its own handler)
        zone.addEventListener('click', (e) => {
            if (!e.target.classList.contains('browse-btn') && !e.target.classList.contains('clear-btn')) {
                browseFile(type);
            }
        });
    });
    
    // Prevent default drag behavior on the window
    window.addEventListener('dragover', (e) => {
        e.preventDefault();
        e.stopPropagation();
    });
    window.addEventListener('drop', (e) => {
        e.preventDefault();
        e.stopPropagation();
    });
}

// Handle file drop
async function handleFileDrop(type, filePath) {
    const zone = type === 'csv' ? csvZone : excelZone;
    const overlay = document.getElementById(`${type}-overlay`);
    
    // Show loading overlay
    overlay.classList.add('active');
    
    try {
        let result;
        if (type === 'csv') {
            result = await window.go.main.App.ValidateCSV(filePath);
        } else {
            result = await window.go.main.App.ValidateExcel(filePath);
        }
        
        // Hide loading overlay
        overlay.classList.remove('active');
        
        if (result.valid) {
            showSuccess(zone, type, result);
            state[type] = {
                valid: true,
                path: result.filePath,
                fileName: result.fileName
            };
        } else {
            showError(zone, type, result.message);
            state[type] = {
                valid: false,
                path: '',
                fileName: ''
            };
        }
        
        updateOkButton();
    } catch (error) {
        overlay.classList.remove('active');
        showError(zone, type, 'Error validating file: ' + error.message);
    }
}

// Show success state
function showSuccess(zone, type, result) {
    zone.classList.remove('invalid', 'shake');
    zone.classList.add('valid');
    
    // Hide drop content, show file info
    zone.querySelector('.drop-zone-content').style.display = 'none';
    
    const fileInfo = document.getElementById(`${type}-info`);
    fileInfo.style.display = 'flex';
    
    document.getElementById(`${type}-filename`).textContent = result.fileName;
    
    const status = document.getElementById(`${type}-status`);
    status.className = 'file-status success';
    status.textContent = '✓ Valid';
    
    document.getElementById(`${type}-details`).innerHTML = 
        `${result.rowCount} rows • ${result.headers.length} columns`;
}

// Show error state
function showError(zone, type, message) {
    zone.classList.remove('valid');
    zone.classList.add('invalid', 'shake');
    
    // Remove shake class after animation
    setTimeout(() => zone.classList.remove('shake'), 400);
    
    // Show file info with error
    zone.querySelector('.drop-zone-content').style.display = 'none';
    
    const fileInfo = document.getElementById(`${type}-info`);
    fileInfo.style.display = 'flex';
    
    document.getElementById(`${type}-filename`).textContent = 'Invalid File';
    
    const status = document.getElementById(`${type}-status`);
    status.className = 'file-status error';
    status.textContent = '✕ Invalid';
    
    document.getElementById(`${type}-details`).innerHTML = message;
}

// Clear file selection
async function clearFile(type) {
    const zone = type === 'csv' ? csvZone : excelZone;
    
    zone.classList.remove('valid', 'invalid');
    zone.querySelector('.drop-zone-content').style.display = 'flex';
    document.getElementById(`${type}-info`).style.display = 'none';
    
    state[type] = {
        valid: false,
        path: '',
        fileName: ''
    };
    
    // Clear in backend
    if (type === 'csv') {
        await window.go.main.App.ClearCSV();
    } else {
        await window.go.main.App.ClearExcel();
    }
    
    updateOkButton();
}

// Browse for file
async function browseFile(type) {
    try {
        const filePath = await window.go.main.App.OpenFileDialog(type);
        if (filePath) {
            await handleFileDrop(type, filePath);
        }
    } catch (error) {
        console.error('Error opening file dialog:', error);
    }
}

// Update OK button state
function updateOkButton() {
    // CSV is required, Excel is optional
    okBtn.disabled = !state.csv.valid;
}

// Process files (OK button)
async function processFiles() {
    if (!state.csv.valid) {
        return;
    }
    
    try {
        const result = await window.go.main.App.ProcessFiles();
        
        if (result.success) {
            // Emit event with file paths for the parent application
            console.log('Processing files:', result);
            // The main application will handle the rest
            window.runtime.Quit();
        } else {
            alert('Error: ' + result.error);
        }
    } catch (error) {
        console.error('Error processing files:', error);
        alert('Error processing files: ' + error.message);
    }
}

// Cancel button
async function cancel() {
    try {
        await window.go.main.App.Cancel();
    } catch (error) {
        // App might already be closing
        console.log('Cancel called');
    }
}

// Initialize when DOM is ready
document.addEventListener('DOMContentLoaded', initDragDrop);

// Also initialize if DOM is already loaded
if (document.readyState !== 'loading') {
    initDragDrop();
}

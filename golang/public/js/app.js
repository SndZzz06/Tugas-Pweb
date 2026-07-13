// Modal management functions (define only if not already provided by page)
if (!window.openModal) {
    window.openModal = function(modalId) {
        const modal = document.getElementById(modalId);
        if (modal) {
            modal.style.display = 'flex';
        }
    }
}

if (!window.closeModal) {
    window.closeModal = function(modalId) {
        const modal = document.getElementById(modalId);
        if (modal) {
            modal.style.display = 'none';
        }
    }
}

if (!window.openDeleteModal) {
    window.openDeleteModal = function(url, itemName) {
        const deleteForm = document.getElementById('deleteModalForm');
        const deleteName = document.getElementById('deleteItemName');
        if (deleteForm && deleteName) {
            deleteForm.action = url;
            deleteName.innerText = itemName;
            openModal('deleteConfirmModal');
        } else {
            const modal = document.getElementById('deleteConfirmModal');
            if (modal) modal.style.display = 'flex';
        }
    }
}

if (!window.closeDeleteModal) {
    window.closeDeleteModal = function() {
        closeModal('deleteModal');
    }
}

if (!window.openEditModal) {
    window.openEditModal = function(dataset) {
        const editTitle = document.getElementById('edit_title');
        if (editTitle) editTitle.value = dataset.taskTitle || '';
        const editDesc = document.getElementById('edit_description');
        if (editDesc) editDesc.value = dataset.taskDescription || '';
        const editStatus = document.getElementById('edit_status');
        if (editStatus) editStatus.value = dataset.taskStatus || 'To Do';
        const editPriority = document.getElementById('edit_priority');
        if (editPriority) editPriority.value = dataset.taskPriority || 'Medium';
        const editDue = document.getElementById('edit_due_date');
        if (editDue) editDue.value = dataset.taskDueDate || '';
        
        const form = document.getElementById('editTaskForm');
        if (form && dataset.taskId) {
            form.action = '/tasks/' + dataset.taskId;
        }
        
        openModal('editTaskModal');
    }
}

if (!window.closeEditModal) {
    window.closeEditModal = function() {
        closeModal('editTaskModal');
    }
}

// Initialize event listeners (run immediately if DOM already loaded)
function initModalEventHandlers() {
    // Close modal when clicking on the background overlay
    const modals = document.querySelectorAll('.modal-overlay');
    modals.forEach(modal => {
        modal.addEventListener('click', function(event) {
            if (event.target === this) {
                this.style.display = 'none';
            }
        });
    });

    // New Task button
    const newTaskBtn = document.getElementById('newTaskBtn');
    if (newTaskBtn) {
        newTaskBtn.addEventListener('click', function() {
            openModal('addTaskModal');
        });
    }

    // Add Team button (on team page)
    const addTeamBtn = document.getElementById('addTeamBtn');
    if (addTeamBtn) {
        addTeamBtn.addEventListener('click', function() {
            openModal('addTeamModal');
        });
    }

    // Close Add Task Modal buttons
    const closeAddTaskModal = document.getElementById('closeAddTaskModal');
    if (closeAddTaskModal) {
        closeAddTaskModal.addEventListener('click', function() {
            closeModal('addTaskModal');
        });
    }

    const cancelAddTaskBtn = document.getElementById('cancelAddTaskBtn');
    if (cancelAddTaskBtn) {
        cancelAddTaskBtn.addEventListener('click', function() {
            closeModal('addTaskModal');
        });
    }

    // Delete Project button
    const deleteProjectBtn = document.getElementById('deleteProjectBtn');
    if (deleteProjectBtn) {
        deleteProjectBtn.addEventListener('click', function() {
            const projectId = window.location.pathname.split('/')[2];
            if (projectId) {
                openDeleteModal('/projects/' + projectId, 'menghapus project ini beserta seluruh task');
            }
        });
    }

    // Edit Task buttons
    const editTaskBtns = document.querySelectorAll('.edit-task-btn');
    editTaskBtns.forEach(btn => {
        btn.addEventListener('click', function() {
            openEditModal(this.dataset);
        });
    });

    // Edit Team buttons
    const editTeamBtns = document.querySelectorAll('.edit-team-btn');
    editTeamBtns.forEach(btn => {
        btn.addEventListener('click', function(event) {
            event.stopPropagation();
            const ds = this.dataset;
            // populate edit team modal
            const name = document.getElementById('edit_team_name');
            const phone = document.getElementById('edit_team_phone');
            const role = document.getElementById('edit_team_role');
            if (name) name.value = ds.teamName || '';
            if (phone) phone.value = ds.teamPhone || '';
            if (role) role.value = ds.teamRole || '';
            const form = document.getElementById('editTeamForm');
            if (form && ds.teamId) form.action = '/team/' + ds.teamId;
            openModal('editTeamModal');
        });
    });

    // Delete Task buttons
    const deleteTaskBtns = document.querySelectorAll('.delete-task-btn');
    console.log('deleteTaskBtns found:', deleteTaskBtns.length);
    deleteTaskBtns.forEach(btn => {
        btn.addEventListener('click', function(event) {
            console.log('delete button clicked', this.dataset);
            event.stopPropagation();
            const taskId = this.dataset.taskId;
            if (taskId) {
                openDeleteModal('/tasks/' + taskId, 'menghapus task ini');
            }
        });
    });

    // Delete Team buttons
    const deleteTeamBtns = document.querySelectorAll('.delete-team-btn');
    deleteTeamBtns.forEach(btn => {
        btn.addEventListener('click', function(event) {
            event.stopPropagation();
            const teamId = this.dataset.teamId;
            if (teamId) {
                openDeleteModal('/team/' + teamId, 'menghapus member ini');
            }
        });
    });

    // Close Edit Task Modal button
    const closeEditTaskModal = document.getElementById('closeEditTaskModal');
    if (closeEditTaskModal) {
        closeEditTaskModal.addEventListener('click', function() {
            closeModal('editTaskModal');
        });
    }

    // data-dismiss handlers (close modal when element clicked)
    const dismissEls = document.querySelectorAll('[data-dismiss="modal"]');
    dismissEls.forEach(el => {
        el.addEventListener('click', function(event) {
            // find closest modal-overlay parent
            let parent = el.closest('.modal-overlay');
            if (parent) parent.style.display = 'none';
        });
    });
}

if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', initModalEventHandlers);
} else {
    // DOM already loaded
    initModalEventHandlers();
}


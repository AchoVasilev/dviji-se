'use strict'

function addErrorToElement(elementId, error = 'Невалидни данни') {
	const inputElement = document.getElementById(elementId);
	if (inputElement.validity.valid) {
		inputElement.classList.remove('input-error');
		removeError(`error-${elementId}`);
		removeClass(`${elementId}-label`, 'error');
		inputElement.setCustomValidity('');

		return;
	}

	inputElement.classList.add('input-error');

	addErrorClass(`${elementId}-label`);
	showError(`error-${elementId}`, error);
	inputElement.setCustomValidity(error);
}

function showError(elementId, error) {
	const element = addErrorClass(elementId);
	element.textContent = error;
	element.classList.add('d-block');
	element.classList.remove('d-none');
}

function addErrorClass(elementId, errorClass = 'error') {
	const element = document.getElementById(elementId);
	element.classList.add(errorClass);

	return element;
}

function checkElementValidity(elementId) {
	const element = document.getElementById(elementId);
	if (element.validity.valid) {
		element.classList.remove('input-error');
		removeError(`error-${elementId}`);
		removeClass(`${elementId}-label`, 'error');
	}

	element.setCustomValidity('');
}

function removeError(elementId) {
	const element = removeClass(elementId, 'd-block');
	element.classList.add('d-none');
}

function removeClass(elementId, classString) {
	const element = document.getElementById(elementId);
	element.classList.remove(classString);

	return element;
}

function checkRepeatPassword() {
	const passwordElement = document.getElementById('password');
	const repeatPasswordElement = document.getElementById('repeat-password');

	if (repeatPasswordElement.value === passwordElement.value) {
		repeatPasswordElement.setCustomValidity('');
		repeatPasswordElement.classList.remove('input-error');
		removeError('error-repeat-password');
		removeClass('repeat-password-label', 'error');

		return;
	}

	repeatPasswordElement.classList.add('input-error');

	addErrorClass(`repeat-password-label`);
	showError(`error-repeat-password`, 'Паролите не съвпадат');
	repeatPasswordElement.setCustomValidity('Паролите не съвпадат');
}

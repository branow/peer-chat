
function createFixedForm(form, button) {
  form.style.display = 'none';
  form.addEventListener('click', (event) => {
    if (event.target === form) {
      form.style.display = 'none';
    }
  });
  button.addEventListener('click', () => {
    form.style.display = 'flex';
  });
}
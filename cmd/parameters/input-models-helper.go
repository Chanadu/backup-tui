package parameters

func (m InputModel) totalItemCount() int {
	return len(m.TextInputs) + len(m.SwitchInputs)
}

func (m InputModel) textInputSelected(indexes ...int) bool {
	if len(indexes) == 0 {
		indexes = append(indexes, m.currentIndex)
	}
	index := indexes[0]

	return index < len(m.TextInputs)
}

func (m InputModel) switchInputSelected(indexes ...int) bool {
	if len(indexes) == 0 {
		indexes = append(indexes, m.currentIndex)
	}
	index := indexes[0]

	return index >= len(m.TextInputs)
}

func (m InputModel) textIndex(indexes ...int) int {
	if len(indexes) == 0 {
		indexes = append(indexes, m.currentIndex)
	}
	index := indexes[0]

	return index
}

func (m InputModel) switchIndex(indexes ...int) int {
	if len(indexes) == 0 {
		indexes = append(indexes, m.currentIndex)
	}
	index := indexes[0]

	return index - len(m.TextInputs)
}

func (m InputModel) blurCurrentIndex() {
	if m.textInputSelected() {
		m.TextInputs[m.textIndex()].Ti.Blur()
	} else if m.switchInputSelected() {
		m.SwitchInputs[m.switchIndex()].Blur()
	}
}

func (m InputModel) focusCurrentIndex() {
	if m.textInputSelected() {
		m.TextInputs[m.textIndex()].Ti.Focus()
	} else if m.switchInputSelected() {
		m.SwitchInputs[m.switchIndex()].Focus()
	}
}

func (m *InputModel) SetCurrentIndex(index int) {
	m.blurCurrentIndex()

	m.currentIndex = index
	m.focusCurrentIndex()
}

func wrap(x, n int) int {
	return ((x % n) + n) % n
}

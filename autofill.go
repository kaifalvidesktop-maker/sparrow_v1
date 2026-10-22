package main

import (
	"encoding/json"
	"sync"
)

type AutofillData struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Phone   string `json:"phone"`
	Address string `json:"address"`
}

var (
	autofillMu   sync.RWMutex
	autofillData AutofillData
)

func getAutofillProfile() AutofillData {
	autofillMu.RLock()
	defer autofillMu.RUnlock()

	return autofillData
}

func saveAutofillProfile(data AutofillData) {
	autofillMu.Lock()
	autofillData = data
	autofillMu.Unlock()
}

func clearAutofillProfile() {
	autofillMu.Lock()
	autofillData = AutofillData{}
	autofillMu.Unlock()
}

func autofillJSON() string {
	autofillMu.RLock()
	defer autofillMu.RUnlock()

	data, err := json.Marshal(autofillData)
	if err != nil {
		return "{}"
	}

	return string(data)
}

const autofillPageHTML = `
<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<title>SPARROW Autofill</title>

<style>
* {
	box-sizing: border-box;
}

body {
	margin: 0;
	padding: 40px 60px;
	background: #0b0b0d;
	color: #e8e8e8;
	font-family: "Segoe UI", sans-serif;
}

body[data-theme="light"] {
	background: #f5f7fa;
	color: #202124;
}

body[data-theme="light"] .card,
body[data-theme="light"] input,
body[data-theme="light"] button {
	background: #ffffff;
	border-color: #dadce0;
	color: #202124;
}

body[data-theme="light"] .subtitle,
body[data-theme="light"] label,
body[data-theme="light"] .status {
	color: #3c4043;
}

body[data-theme="light"] .status {
	color: #188038;
}

body[data-theme="light"] button.danger {
	color: #d93025;
	border-color: #f1b5b5;
}

body[data-theme="light"] button.danger:hover {
	background: #fce8e6;
}

h1 {
	margin: 0 0 8px;
	font-size: 26px;
	color: #f4f4f4;
}

.subtitle {
	color: #a0a0a0;
	font-size: 13px;
	margin-bottom: 28px;
}

.card {
	max-width: 720px;
	background: #17181a;
	border: 1px solid #232323;
	border-radius: 14px;
	padding: 24px;
	box-shadow: 0 2px 8px rgba(0,0,0,.24);
}

.field {
	margin-bottom: 18px;
}

label {
	display: block;
	font-size: 12px;
	font-weight: 600;
	margin-bottom: 7px;
	color: #d7d7d7;
}

input {
	width: 100%;
	padding: 11px 13px;
	border: 1px solid #343434;
	border-radius: 8px;
	background: #111315;
	color: #f0f0f0;
	outline: none;
	font-size: 13px;
}

input:focus {
	border-color: #4d90fe;
}

.buttons {
	display: flex;
	gap: 10px;
	margin-top: 24px;
}

button {
	border: 1px solid #353535;
	border-radius: 8px;
	padding: 10px 18px;
	background: #111315;
	color: #eaeaea;
	cursor: pointer;
	font-size: 12px;
}

button:hover {
	background: #1a1d21;
}

button.primary {
	background: #1967d2;
	border-color: #1967d2;
	color: #fff;
}

button.primary:hover {
	background: #185abc;
}

button.danger {
	color: #ff8a80;
	border-color: #5a2d2d;
}

button.danger:hover {
	background: #2a1b1a;
}

.status {
	margin-top: 16px;
	font-size: 12px;
	color: #7fe3a0;
}
</style>
</head>

<body>

<h1>Autofill</h1>

<div class="subtitle">
	Save basic information for faster form filling.
</div>

<div class="card">

	<div class="field">
		<label>Name</label>
		<input id="name" type="text" autocomplete="name">
	</div>

	<div class="field">
		<label>Email</label>
		<input id="email" type="email" autocomplete="email">
	</div>

	<div class="field">
		<label>Phone</label>
		<input id="phone" type="tel" autocomplete="tel">
	</div>

	<div class="field">
		<label>Address</label>
		<input id="address" type="text" autocomplete="street-address">
	</div>

	<div class="buttons">
		<button class="primary" onclick="saveData()">
			Save
		</button>

		<button class="danger" onclick="clearData()">
			Clear
		</button>
	</div>

	<div id="status" class="status"></div>

</div>

<script>

async function loadData() {
	try {
		const data = await window.getAutofillProfile();

		if (!data) {
			return;
		}

		document.getElementById("name").value =
			data.name || "";

		document.getElementById("email").value =
			data.email || "";

		document.getElementById("phone").value =
			data.phone || "";

		document.getElementById("address").value =
			data.address || "";

	} catch (error) {
		console.error(error);
	}
}

async function saveData() {

	const data = {
		name:
			document.getElementById("name").value,

		email:
			document.getElementById("email").value,

		phone:
			document.getElementById("phone").value,

		address:
			document.getElementById("address").value
	};

	try {

		await window.saveAutofillProfile(data);

		document.getElementById("status").textContent =
			"Autofill information saved.";

	} catch (error) {

		document.getElementById("status").textContent =
			"Could not save autofill information.";
	}
}

async function clearData() {

	try {

		await window.clearAutofillProfile();

		document.getElementById("name").value = "";
		document.getElementById("email").value = "";
		document.getElementById("phone").value = "";
		document.getElementById("address").value = "";

		document.getElementById("status").textContent =
			"Autofill information cleared.";

	} catch (error) {

		document.getElementById("status").textContent =
			"Could not clear autofill information.";
	}
}

loadData();

</script>

</body>
</html>
`

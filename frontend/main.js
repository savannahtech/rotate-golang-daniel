import { Start, Stop } from './wailsjs/go/core/App.js';

document.getElementById("startButton").addEventListener("click", function () {
    Start().then(function () {
        document.getElementById("status").innerText = "App started successfully!";
        document.getElementById("stopButton").disabled = false;
        document.getElementById("startButton").disabled = true;
        document.getElementById("fetchLogsButton").disabled = false;
    }).catch(function (err) {
        document.getElementById("status").innerText = "Error: " + err;
    });
});

document.getElementById("stopButton").addEventListener("click", function () {
    Stop().then(function () {
        document.getElementById("status").innerText = "App stopped successfully!";
        document.getElementById("startButton").disabled = false;
        document.getElementById("fetchLogsButton").disabled = true;
        document.getElementById("stopButton").disabled = true;
    }).catch(function (err) {
        document.getElementById("status").innerText = "Error: " + err;
    });
});

document.getElementById("fetchLogsButton").addEventListener("click", function () {
    fetch("http://localhost:9000/v1/logs?limit=2")
        .then(response => response.json())
        .then(data => {
            console.log(data)
            displayLogsInTable(data);
        })
        .catch(err => {
            document.getElementById("status").innerText = "Error fetching logs: " + err;
        });
});

function displayLogsInTable(logs) {
    console.log(logs)
    const logsTableBody = document.getElementById("logsTableBody");
    logsTableBody.innerHTML = '';  // Clear existing rows

    logs.forEach(log => {
        const row = document.createElement("tr");

        const idCell = document.createElement("td");
        idCell.innerText = log.id;
        row.appendChild(idCell);

        const detailsCell = document.createElement("td");
        const pre = document.createElement("pre");
        pre.innerText = JSON.stringify(log.details, null, 2);  // JSON string with indentation
        detailsCell.appendChild(pre);
        row.appendChild(detailsCell);

        const logTimeCell = document.createElement("td");
        logTimeCell.innerText = log.logTime;
        row.appendChild(logTimeCell);

        logsTableBody.appendChild(row);
    });
}


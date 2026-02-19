const chartDiv = document.getElementById("chart");

// Data для uPlot: [ [time], [avg], [min], [max], [last] ]
let data = [
    [],
    [],
    [],
    [],
    []
];


const opts = {
    title: "BTC/USDT 5s Moving Average",
    width: 800,
    height: 400,
    series: [{}, // 0: X (Time)
        {
            label: "Average",
            stroke: "#40fcf4",
            width: 1,
            points: {
                show: false
            },
        },
        {
            label: "Min",
            stroke: "transparent",
            width: 2,
            points: {
                show: false
            },
            spanGaps: true
        },
        {
            label: "Max",
            stroke: "transparent",
            width: 1,
            points: {
                show: false
            },
            spanGaps: true
        },
        {
            label: "Live",
            stroke: "rgba(255,255,255,0.5)",
            width: 1,
            points: {
                show: false
            },
        },
    ],
    bands: [{
            series: [1, 2],
            fill: "rgba(64, 252, 244, 0.07)",
        },
        {
            series: [3, 1],
            fill: "rgba(64, 252, 244, 0.07)",
        }
    ],
    scales: {
        x: {
            time: true
        }
    },
    axes: [{}, {
        space: 40
    }],
    width: window.innerWidth - 40,
};

let uplot = new uPlot(opts, data, chartDiv);

const ws = new WebSocket("ws://localhost:8080/ws?format=json");

ws.onmessage = (event) => {
    const statsArray = JSON.parse(event.data);
    const btc = statsArray.find(s => s.s === "BTCUSDT");

    if (btc) {
        const now = Math.floor(Date.now() / 1000);

        data[0].push(now);
        data[1].push(btc.a);
        data[2].push(btc.m);
        data[3].push(btc.x);
        data[4].push(btc.p);

        // Take last X points on chart
        if (data[0].length > 1000) {
            data[0].shift();
            data[1].shift();
            data[2].shift();
            data[3].shift();
            data[4].shift();
        }

        uplot.setData(data);
    }
};

ws.onopen = () => console.log("Connected to ws");
ws.onerror = (e) => console.error("WS Error", e);

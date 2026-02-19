const symbols = ["BTCUSDT", "ETHUSDT", "SOLUSDT"];

const root = document.getElementById("root");


const getOptions = (s) => ({
    title: `${s} Moving Average`,
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
});

const getChartDiv = (s) => document.getElementById(`chart-${s}`);


const plotFactory = (symbol) => {
    // Data для uPlot: [ [time], [avg], [min], [max], [last] ]
    let data = [
        [],
        [],
        [],
        [],
        []
    ];

    let uplot = new uPlot(getOptions(symbol), data, getChartDiv(symbol));

    return {
        uplot,
        data
    };
};

const getCharts = () => symbols.reduce((acc, cur) => ({
    ...acc,
    [cur]: plotFactory(cur)
}), {});


const ws = new WebSocket("ws://localhost:8080/ws?format=json");


symbols.forEach((s) => {
    const chartContainer = document.createElement('div');
    chartContainer.id = `chart-${s}`;
    chartContainer.classList.add('chart-container');
    root.append(chartContainer);
});

const charts = getCharts();

console.log(charts);

ws.onmessage = (event) => {
    const statsArray = JSON.parse(event.data);

    statsArray.forEach((item) => {
        if (!charts[item.s]) return;

        const {
            data,
            uplot
        } = charts[item.s];

        console.log(item.s, charts[item.s].data);
        const now = Math.floor(Date.now() / 1000);

        data[0].push(now);
        data[1].push(item.a);
        data[2].push(item.m);
        data[3].push(item.x);
        data[4].push(item.p);

        // Take last X points on chart
        if (data[0].length > 1000) {
            data[0].shift();
            data[1].shift();
            data[2].shift();
            data[3].shift();
            data[4].shift();
        }

        uplot.setData(data);
    });
};

ws.onopen = () => console.log("Connected to ws");
ws.onerror = (e) => console.error("WS Error", e);

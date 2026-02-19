const symbols = [
    "BTCUSDT",
    "ETHUSDT",
    "SOLUSDT",
    "BNBUSDT",
    "XRPUSDT",
    "ADAUSDT",
    "DOGEUSDT",
    "DOTUSDT",
    "AVAXUSDT",
    "TRXUSDT",
    "LTCUSDT",
    "LINKUSDT",
    "NEARUSDT",
    "ATOMUSDT",
    "ARBUSDT",
    "OPUSDT",
    "LDOUSDT",
];

const getOptions = (s) => ({
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
    width: 400,
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

const dashboard = document.getElementById('dashboard');

symbols.forEach((s) => {
    const wrapper = document.createElement('div');
    wrapper.classList.add('chart-container');
    wrapper.id = `wrapper-${s}`;

    const title = document.createElement('div');
    title.classList.add('chart-title');
    title.innerText = s;
    wrapper.appendChild(title);

    const chartDiv = document.createElement('div');
    chartDiv.id = `chart-${s}`;
    wrapper.appendChild(chartDiv);

    dashboard.appendChild(wrapper);
});


const charts = getCharts();

const rootProto = protobuf.Root.fromJSON({
    nested: {
        models: {
            nested: {
                TradeStatsProto: {
                    fields: {
                        symbol: {
                            id: 1,
                            type: "string"
                        },
                        price: {
                            id: 2,
                            type: "double"
                        },
                        avgPrice: {
                            id: 3,
                            type: "double"
                        },
                        minPrice: {
                            id: 4,
                            type: "double"
                        },
                        maxPrice: {
                            id: 5,
                            type: "double"
                        }
                    }
                },
                MarketUpdateProto: {
                    fields: {
                        stats: {
                            id: 1,
                            rule: "repeated",
                            type: "TradeStatsProto"
                        }
                    }
                }
            }
        }
    }
});

const MarketUpdate = rootProto.lookupType("models.MarketUpdateProto");

const ws = new WebSocket("ws://localhost:8080/ws?format=proto");
ws.binaryType = "arraybuffer";

ws.onmessage = (event) => {
    const uint8view = new Uint8Array(event.data);
    const message = MarketUpdate.decode(uint8view);

    const {
        stats
    } = MarketUpdate.toObject(message);

    if (!stats) return;

    stats.forEach((item) => {
        if (!charts[item.symbol]) return;

        const vol = ((item.maxPrice - item.minPrice) / item.avgPrice) * 100;

        const titleEl = document.querySelector(`#wrapper-${item.symbol} .chart-title`);

        if (titleEl) {
            titleEl.innerText = `${item.symbol} | Vol: ${vol.toFixed(3)}%`;
        }

        const wrapper = document.getElementById(`wrapper-${item.symbol}`);
        if (wrapper) {
            wrapper.style.order = Math.round(vol * -1000);
        }

        const {
            data,
            uplot
        } = charts[item.symbol];
        const now = Math.floor(Date.now() / 1000);

        data[0].push(now);
        data[1].push(item.avgPrice);
        data[2].push(item.minPrice);
        data[3].push(item.maxPrice);
        data[4].push(item.price);

        if (data[0].length > 200) {
            data.forEach(arr => arr.shift());
        }

        uplot.setData(data);
    });
};

ws.onopen = () => console.log("Connected to ws");
ws.onerror = (e) => console.error("WS Error", e);

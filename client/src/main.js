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
        market: {
            nested: {
                v1: {
                    nested: {
                        TickStatsProto: {
                            fields: {
                                Symbol: {
                                    id: 1,
                                    type: "string"
                                },
                                Price: {
                                    id: 2,
                                    type: "string"
                                },
                                AvgPrice: {
                                    id: 3,
                                    type: "string"
                                },
                                MinPrice: {
                                    id: 4,
                                    type: "string"
                                },
                                MaxPrice: {
                                    id: 5,
                                    type: "string"
                                },
                                IsVolumeSpike: {
                                    id: 6,
                                    type: "bool"
                                }

                            }
                        },
                        MarketSnapshot: {
                            fields: {
                                stats: {
                                    id: 1,
                                    rule: "repeated",
                                    type: "TickStatsProto"
                                }
                            }
                        }
                    }
                }
            }
        }
    }
});


const MarketSnapshot = rootProto.lookupType("market.v1.MarketSnapshot");

const ws = new WebSocket("ws://localhost:8080/ws?format=proto");
ws.binaryType = "arraybuffer";

ws.onmessage = (event) => {
    try {
        const uint8view = new Uint8Array(event.data);


        const message = MarketSnapshot.decode(uint8view);
        const decoded = MarketSnapshot.toObject(message, {
            longs: String,
            enums: String,
            bytes: String,
            defaults: true
        });

        const stats = decoded.stats;
        if (!stats || !Array.isArray(stats)) return;

        stats.forEach((item) => {

            if (!charts[item.Symbol.toUpperCase()]) return;

            const price = parseFloat(item.Price);
            const avg = parseFloat(item.AvgPrice);
            const min = parseFloat(item.MinPrice);
            const max = parseFloat(item.MaxPrice);


            const vol = ((max - min) / avg) * 100;

            const titleEl = document.querySelector(`#wrapper-${item.Symbol} .chart-title`);
            if (titleEl) {
                // FIXME
                const spikeIcon = item.IsVolumeSpike ? "🔥" : "";
                titleEl.innerText = `${spikeIcon}${item.Symbol} | Vol: ${vol.toFixed(3)}%`;
            }

            const wrapper = document.getElementById(`wrapper-${item.Symbol}`);
            if (wrapper) {
                wrapper.style.order = Math.round(vol * -1000);

                wrapper.classList.toggle('volume-spike', item.IsVolumeSpike);
            }

            const { data, uplot } = charts[item.Symbol.toUpperCase()];
            const now = Math.floor(Date.now() / 1000);

            data[0].push(now);
            data[1].push(avg);
            data[2].push(min);
            data[3].push(max);
            data[4].push(price);

            if (data[0].length > 200) {
                data.forEach(arr => arr.shift());
            }

            uplot.setData(data);
        });
    } catch (e) {
        console.error("Decode error:", e);
    }
};

ws.onopen = () => console.log("Connected to ws");
ws.onerror = (e) => console.error("WS Error", e);

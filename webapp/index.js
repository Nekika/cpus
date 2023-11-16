import { h, render } from 'https://esm.sh/preact';
import { useEffect, useState } from 'https://esm.sh/preact/hooks';
import htm from 'https://esm.sh/htm';

// Initialize htm with Preact
const html = htm.bind(h);

function Usage(props) {
    const name = `cpu${props.id}`
    const value = props.value.toFixed(2)

    const style = { display: 'flex', gap: '12px'}

    return html`<div class="usage" style="${style}">
        <span>${name}</span>
        <span>${value} %</span>
    </div>`
}

function App() {
    const [usages, setUsages] = useState([])

    useEffect(() => {
        const url = new URL('/ws/cpus', document.URL)
        url.protocol = url.protocol.replace('http', 'ws')

        const conn = new WebSocket(url)

        conn.addEventListener('message', (event) => {
            setUsages(JSON.parse(event.data))
        })

        return () => {
            conn.close()
        }
    }, [])

    return html`<div>
        <div class="usages">
            ${usages.map((usage, index) => html`<${Usage} id="${index}" value="${usage}"/>`)}
        </div>
    </div>`
}

document.addEventListener('DOMContentLoaded', () => {
    render(html`<${App} />`, document.body);
})
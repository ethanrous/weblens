let url = ''

const getData = async () => {
    const res = await fetch(url)
        .then((data) => data.json())
        .then((data) => data.data)
    return res
}

const Test = () => {
    return <p style={{ textAlign: 'center', marginTop: '27%' }}>no</p>
}

export default Test

config = require("./helper")

module.exports.list = async (client, ns) => {
    try {
        await client.loadSpec();
        const secret = await client.api.v1.namespaces(ns).secrets().get()
        console.log('Secrets: ', secret)

    } catch (err) {
        console.error('Error: ', err)
    }

}


module.exports.create = async (client, ns, secretManifest) => {
    try {
        await client.loadSpec();

        const secret = await client.api.v1.namespaces(ns).secrets.post({
            body: secretManifest
        })
        console.log('Create:', secret.body)

    } catch (err) {
        console.error('Error: ', err)
    }

}

module.exports.delete = async (client, ns, name) => {
    try {
        await client.loadSpec();

        const secret = await client.api.v1.namespaces(ns).secrets(name).delete()
        console.log('Removed: ', secret)

    } catch (err) {
        console.error('Error: ', err)
    }

}
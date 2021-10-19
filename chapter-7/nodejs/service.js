config = require("./helper")

module.exports.list = async (client, ns) => {
    try {
        await client.loadSpec();
        const svc = await client.api.v1.namespaces(ns).services().get()
        console.log('Services: ', svc)

    } catch (err) {
        console.error('Error: ', err)
    }

}


module.exports.create = async (client, ns, svcManifest) => {
    try {
        await client.loadSpec();

        const create = await client.api.v1.namespaces(ns).services.post({
            body: svcManifest
        })
        console.log('Create:', create.body)

    } catch (err) {
        console.error('Error: ', err)
    }

}

module.exports.delete = async (client, ns, name) => {
    try {
        await client.loadSpec();

        const removed = await client.api.v1.namespaces(ns).services(name).delete()
        console.log('Removed: ', removed)

    } catch (err) {
        console.error('Error: ', err)
    }

}
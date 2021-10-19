config = require("./helper")

module.exports.list = async (client, ns) => {
    try {
        await client.loadSpec();
        const pvc = await client.api.v1.namespaces(ns).pvc().get()
        console.log('PVC: ', pvc)

    } catch (err) {
        console.error('Error: ', err)
    }

}


module.exports.create = async (client, ns, pvcManifest) => {
    try {
        await client.loadSpec();

        const create = await client.api.v1.namespaces(ns).pvc.post({
            body: pvcManifest
        })
        console.log('Create:', create.body)

    } catch (err) {
        console.error('Error: ', err)
    }

}

module.exports.delete = async (client, ns, name) => {
    try {
        await client.loadSpec();

        const removed = await client.api.v1.namespaces(ns).pvc(name).delete()
        console.log('Removed: ', removed)

    } catch (err) {
        console.error('Error: ', err)
    }

}
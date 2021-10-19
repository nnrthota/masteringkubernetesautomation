config = require("./helper")

module.exports.list = async (client, ns) => {
    try {
        await client.loadSpec();
        const deployment = await client.apis.apps.v1.namespaces(ns).deployments().get()
        console.log('Deployment: ', deployment)

    } catch (err) {
        console.error('Error: ', err)
    }

}


module.exports.create = async (client, ns, deploymentManifest) => {
    try {
        await client.loadSpec();

        const create = await client.apis.apps.v1.namespaces(ns).deployments.post({
            body: deploymentManifest
        })
        console.log('Create:', create.body)

    } catch (err) {
        console.error('Error: ', err)
    }

}

module.exports.delete = async (client, ns, name) => {
    try {
        await client.loadSpec();

        const removed = await client.apis.apps.v1.namespaces(ns).deployments(name).delete()
        console.log('Removed: ', removed)

    } catch (err) {
        console.error('Error: ', err)
    }

}
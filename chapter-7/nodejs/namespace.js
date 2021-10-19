config = require("./helper")

module.exports.list = async (client) =>{
    try {
        client = config.Client(await config.getConfigFile())
        await client.loadSpec();
        const namespaces = await client.api.v1.namespaces.get()
        console.log('Namespaces: ', namespaces)

    } catch (err) {
        console.error('Error: ', err)
    }

}
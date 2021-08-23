import { InjectionKey } from '@nuxtjs/composition-api'
import { NodeStore } from '../../store/node'

const NodeStoreKey: InjectionKey<NodeStore> = Symbol('NodeStore')
export default NodeStoreKey

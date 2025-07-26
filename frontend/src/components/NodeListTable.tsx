import { FaServer as ServerCog } from "react-icons/fa"
import { Card } from "@/components/ui/card"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import ClientIcon from "./ClientIcon"

interface NodeInfo {
  name: string
  type: string
  server: string
  port: number
}

interface NodeListTableProps {
  nodes: NodeInfo[]
  filteredNodes: NodeInfo[]
  testing: boolean
}

export default function NodeListTable({
  nodes,
  filteredNodes,
  testing,
}: NodeListTableProps) {
  return (
    <Card className="card-elevated">
      <div className="flex items-center justify-between form-element">
        <h2 className="text-lg font-semibold text-lavender-50 flex items-center gap-2">
          <ClientIcon icon={ServerCog} className="h-5 w-5 text-lavender-400" />
          节点列表 {testing ? "(测试中)" : "(预览)"}
        </h2>
      </div>

      <div className="table-container scrollbar-modern">
        <Table className="table-modern">
          <TableHeader
            style={{ position: "sticky", top: 0, zIndex: 10, backdropFilter: "blur(8px)" }}
          >
            <TableRow>
              <TableHead>名称</TableHead>
              <TableHead>协议</TableHead>
              <TableHead>IP / 域名</TableHead>
              <TableHead>端口</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {filteredNodes.length > 0 ? (
              filteredNodes.map((node, index) => (
                <TableRow key={`${node.name}-${index}`}>
                  <TableCell className="font-medium text-lavender-50">
                    <div className="truncate max-w-xs" title={node.name}>
                      {node.name}
                    </div>
                  </TableCell>
                  <TableCell>
                    <span className={`badge-filled protocol-${node.type.toLowerCase()}`}>
                      {node.type}
                    </span>
                  </TableCell>
                  <TableCell className="text-lavender-300 font-mono text-sm">
                    {node.server}
                  </TableCell>
                  <TableCell className="text-lavender-300">{node.port}</TableCell>
                </TableRow>
              ))
            ) : (
              <TableRow>
                <TableCell colSpan={4} className="text-center text-lavender-400 py-8">
                  {nodes.length === 0 ? "暂无节点信息" : "没有符合条件的节点"}
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>
    </Card>
  )
}
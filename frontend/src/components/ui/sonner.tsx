import { Toaster as Sonner } from "sonner"

const Toaster = ({ ...props }: React.ComponentProps<typeof Sonner>) => {
  return (
    <Sonner
      className="toaster group"
      toastOptions={{
        classNames: {
          toast:
            "group toast group-[.toaster]:bg-shamrock-900 group-[.toaster]:text-shamrock-50 group-[.toaster]:border-shamrock-500 group-[.toaster]:shadow-lg",
          description: "group-[.toast]:text-shamrock-300",
          actionButton:
            "group-[.toast]:bg-shamrock-600 group-[.toast]:text-shamrock-50 group-[.toast]:hover:bg-shamrock-700",
          cancelButton:
            "group-[.toast]:bg-shamrock-800 group-[.toast]:text-shamrock-200 group-[.toast]:hover:bg-shamrock-700",
          closeButton:
            "group-[.toast]:bg-shamrock-800 group-[.toast]:text-shamrock-200 group-[.toast]:hover:bg-shamrock-700",
        },
      }}
      {...props}
    />
  )
}

export { Toaster }

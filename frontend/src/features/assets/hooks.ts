import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { api } from "../../lib/api/client";

export function useAssets(search = "") {
  return useQuery({
    queryKey: ["assets", search],
    queryFn: () => api.assets(search)
  });
}

export function useAssetMutations() {
  const queryClient = useQueryClient();

  return {
    create: useMutation({
      mutationFn: api.createAsset,
      onSuccess: async () => {
        await queryClient.invalidateQueries({ queryKey: ["assets"] });
      }
    })
  };
}

export function useAssetIcon(assetId?: string) {
  return useQuery({
    queryKey: ["asset-icon", assetId],
    queryFn: () => api.assetIcon(assetId ?? ""),
    enabled: Boolean(assetId)
  });
}

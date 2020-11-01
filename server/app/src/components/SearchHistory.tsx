import { Dispatch } from "@reduxjs/toolkit";
import { Badge, EyeOffIcon, EyeOpenIcon, Menu, Pane, Popover, Position, SearchIcon, Spinner, Text, TrashIcon } from 'evergreen-ui';
import React from 'react';
import { useDispatch, useSelector } from "react-redux";
import { deleteHistoryItem, HistoryItem, selectHistory } from "../state/historySlice";
import { setActiveItem } from "../state/stateSlice";
import { RootState } from "../state/store";


const SearchHistory: React.FC = () => {
    const history = useSelector(selectHistory);
    const activeTS = useSelector((store: RootState) => store.state.activeItem?.timestamp) ?? -1;
    const dispatch = useDispatch();

    return (
        <>
            {
                history.length > 0 ? history.map((item: HistoryItem) =>
                    <HistoryCard
                        activeTS={activeTS}
                        key={item.timestamp.toString()}
                        item={item}
                        dispatch={dispatch} />)
                    : <Pane display="flex" justifyContent="center">
                        <Text marginX="auto" marginY={16} color="muted">History is a mystery.</Text>
                    </Pane >
            }
        </>
    );
}

type Props = {
    activeTS: number;
    item: HistoryItem;
    dispatch: Dispatch<any>;
}

const HistoryCard: React.FC<Props> = ({ activeTS, item, dispatch }: Props) => {
    return (
        <Popover
            position={Position.BOTTOM}
            content={
                <Menu>
                    {activeTS !== item.timestamp &&
                        <Menu.Item
                            icon={EyeOpenIcon}
                            onClick={() => dispatch(setActiveItem(item))}>
                            Show
                            </Menu.Item>
                    }
                    {activeTS === item.timestamp &&
                        <Menu.Item
                            icon={EyeOffIcon}
                            onClick={() => dispatch(setActiveItem(null))}>
                            Hide
                            </Menu.Item>
                    }
                    <Menu.Item icon={TrashIcon}
                        onClick={() => dispatch(deleteHistoryItem(item.timestamp))}
                        intent="danger">
                        Delete
                        </Menu.Item>
                </Menu>
            }>
            <Pane
                cursor="pointer"
                border
                padding={6}
                elevation={1}
                margin={8}
                display="flex"
                justifyContent="space-between"
                alignItems="center"
            >
                <SearchIcon size={15} color="#234361" />
                <Text width={140}
                    marginLeft={24}
                    flex={1}
                    whiteSpace="nowrap"
                    display="block"
                    fontWeight={500}
                    overflow="hidden"
                    textOverflow="ellipsis">
                    {item.query}
                </Text>
                {!item.results?.length ?
                    <Spinner size={24} marginRight={5} /> :
                    <Badge color="green">
                        {`${item.results?.length} RESULTS`}
                    </Badge>
                }
            </Pane>
        </Popover>
    )
}

export default SearchHistory;